// Copyright Cloud Ridge Works 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"net/http"

	"github.com/cloudridgeworks/terraform-provider-revenuecat/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &OfferingResource{}
var _ resource.ResourceWithConfigure = &OfferingResource{}
var _ resource.ResourceWithImportState = &OfferingResource{}

type OfferingResource struct{ client *client.Client }

type OfferingResourceModel struct {
	ProjectID   types.String `tfsdk:"project_id"`
	ID          types.String `tfsdk:"id"`
	LookupKey   types.String `tfsdk:"lookup_key"`
	DisplayName types.String `tfsdk:"display_name"`
	IsCurrent   types.Bool   `tfsdk:"is_current"`
	State       types.String `tfsdk:"state"`
	CreatedAt   types.Int64  `tfsdk:"created_at"`
}

type offeringResponse struct {
	ID          string `json:"id"`
	LookupKey   string `json:"lookup_key"`
	DisplayName string `json:"display_name"`
	IsCurrent   bool   `json:"is_current"`
	State       string `json:"state"`
	CreatedAt   int64  `json:"created_at"`
}

func NewOfferingResource() resource.Resource { return &OfferingResource{} }

func (r *OfferingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_offering"
}

func (r *OfferingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "A RevenueCat offering. RevenueCat offering metadata is not currently managed by this resource.", Attributes: map[string]schema.Attribute{
		"project_id":   schema.StringAttribute{Required: true, MarkdownDescription: "RevenueCat project ID (1-255 characters).", Validators: identifierValidators(), PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"id":           schema.StringAttribute{Computed: true, MarkdownDescription: "RevenueCat offering ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"lookup_key":   schema.StringAttribute{Required: true, MarkdownDescription: "Stable offering lookup key (1-200 characters).", Validators: lookupKeyValidators(), PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"display_name": schema.StringAttribute{Required: true, MarkdownDescription: "Human-readable offering name (1-1500 characters).", Validators: displayNameValidators()},
		"is_current":   schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Whether this is the project's current offering."},
		"state":        schema.StringAttribute{Computed: true, MarkdownDescription: "RevenueCat lifecycle state."},
		"created_at":   schema.Int64Attribute{Computed: true, MarkdownDescription: "Creation time in milliseconds since Unix epoch."},
	}}
}

func (r *OfferingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureClient(req, resp, &r.client)
}

func (r *OfferingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OfferingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var remote offeringResponse
	err := r.client.Do(ctx, http.MethodPost, escaped("projects", data.ProjectID.ValueString(), "offerings"), map[string]any{"lookup_key": data.LookupKey.ValueString(), "display_name": data.DisplayName.ValueString()}, &remote)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create RevenueCat offering", err.Error())
		return
	}
	if data.IsCurrent.ValueBool() {
		err = r.client.Do(ctx, http.MethodPost, escaped("projects", data.ProjectID.ValueString(), "offerings", remote.ID), map[string]any{"is_current": true}, &remote)
		if err != nil {
			resp.Diagnostics.AddError("Unable to make RevenueCat offering current", err.Error())
			return
		}
	}
	setOfferingState(&data, remote)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OfferingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OfferingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var remote offeringResponse
	err := r.client.Do(ctx, http.MethodGet, escaped("projects", data.ProjectID.ValueString(), "offerings", data.ID.ValueString()), nil, &remote)
	if readError(err, ctx, resp, "offering") {
		return
	}
	setOfferingState(&data, remote)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OfferingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OfferingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var remote offeringResponse
	err := r.client.Do(ctx, http.MethodPost, escaped("projects", data.ProjectID.ValueString(), "offerings", data.ID.ValueString()), map[string]any{"display_name": data.DisplayName.ValueString(), "is_current": data.IsCurrent.ValueBool()}, &remote)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update RevenueCat offering", err.Error())
		return
	}
	setOfferingState(&data, remote)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OfferingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OfferingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Do(ctx, http.MethodDelete, escaped("projects", data.ProjectID.ValueString(), "offerings", data.ID.ValueString()), nil, nil)
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to delete RevenueCat offering", err.Error())
	}
}

func (r *OfferingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importTwoPartID(ctx, req, resp)
}

func setOfferingState(data *OfferingResourceModel, remote offeringResponse) {
	data.ID = types.StringValue(remote.ID)
	data.LookupKey = types.StringValue(remote.LookupKey)
	data.DisplayName = types.StringValue(remote.DisplayName)
	data.IsCurrent = types.BoolValue(remote.IsCurrent)
	data.State = types.StringValue(remote.State)
	data.CreatedAt = types.Int64Value(remote.CreatedAt)
}
