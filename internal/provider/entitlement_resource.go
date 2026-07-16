// Copyright Cloud Ridge Works
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"net/http"

	"github.com/cloudridgeworks/terraform-provider-revenuecat/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &EntitlementResource{}
var _ resource.ResourceWithConfigure = &EntitlementResource{}
var _ resource.ResourceWithImportState = &EntitlementResource{}

type EntitlementResource struct{ client *client.Client }

type EntitlementResourceModel struct {
	ProjectID   types.String `tfsdk:"project_id"`
	ID          types.String `tfsdk:"id"`
	LookupKey   types.String `tfsdk:"lookup_key"`
	DisplayName types.String `tfsdk:"display_name"`
	State       types.String `tfsdk:"state"`
	CreatedAt   types.Int64  `tfsdk:"created_at"`
}

type entitlementResponse struct {
	ID          string `json:"id"`
	LookupKey   string `json:"lookup_key"`
	DisplayName string `json:"display_name"`
	State       string `json:"state"`
	CreatedAt   int64  `json:"created_at"`
}

func NewEntitlementResource() resource.Resource { return &EntitlementResource{} }

func (r *EntitlementResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entitlement"
}

func (r *EntitlementResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "A RevenueCat entitlement.", Attributes: map[string]schema.Attribute{
		"project_id":   schema.StringAttribute{Required: true, MarkdownDescription: "RevenueCat project ID (1-255 characters).", Validators: identifierValidators(), PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"id":           schema.StringAttribute{Computed: true, MarkdownDescription: "RevenueCat entitlement ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"lookup_key":   schema.StringAttribute{Required: true, MarkdownDescription: "Stable entitlement lookup key (1-200 characters).", Validators: lookupKeyValidators(), PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"display_name": schema.StringAttribute{Required: true, MarkdownDescription: "Human-readable entitlement name (1-1500 characters).", Validators: displayNameValidators()},
		"state":        schema.StringAttribute{Computed: true, MarkdownDescription: "RevenueCat lifecycle state."},
		"created_at":   schema.Int64Attribute{Computed: true, MarkdownDescription: "Creation time in milliseconds since Unix epoch."},
	}}
}

func (r *EntitlementResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureClient(req, resp, &r.client)
}

func (r *EntitlementResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EntitlementResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var remote entitlementResponse
	err := r.client.Do(ctx, http.MethodPost, escaped("projects", data.ProjectID.ValueString(), "entitlements"), map[string]any{"lookup_key": data.LookupKey.ValueString(), "display_name": data.DisplayName.ValueString()}, &remote)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create RevenueCat entitlement", err.Error())
		return
	}
	setEntitlementState(&data, remote)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EntitlementResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EntitlementResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var remote entitlementResponse
	err := r.client.Do(ctx, http.MethodGet, escaped("projects", data.ProjectID.ValueString(), "entitlements", data.ID.ValueString()), nil, &remote)
	if readError(err, ctx, resp, "entitlement") {
		return
	}
	setEntitlementState(&data, remote)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EntitlementResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EntitlementResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var remote entitlementResponse
	err := r.client.Do(ctx, http.MethodPost, escaped("projects", data.ProjectID.ValueString(), "entitlements", data.ID.ValueString()), map[string]any{"display_name": data.DisplayName.ValueString()}, &remote)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update RevenueCat entitlement", err.Error())
		return
	}
	setEntitlementState(&data, remote)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EntitlementResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EntitlementResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Do(ctx, http.MethodDelete, escaped("projects", data.ProjectID.ValueString(), "entitlements", data.ID.ValueString()), nil, nil)
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to delete RevenueCat entitlement", err.Error())
	}
}

func (r *EntitlementResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importTwoPartID(ctx, req, resp)
}

func setEntitlementState(data *EntitlementResourceModel, remote entitlementResponse) {
	data.ID = types.StringValue(remote.ID)
	data.LookupKey = types.StringValue(remote.LookupKey)
	data.DisplayName = types.StringValue(remote.DisplayName)
	data.State = types.StringValue(remote.State)
	data.CreatedAt = types.Int64Value(remote.CreatedAt)
}
