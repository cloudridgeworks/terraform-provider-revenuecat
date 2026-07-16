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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &PackageResource{}
var _ resource.ResourceWithConfigure = &PackageResource{}
var _ resource.ResourceWithImportState = &PackageResource{}

type PackageResource struct{ client *client.Client }

type PackageResourceModel struct {
	ProjectID   types.String `tfsdk:"project_id"`
	OfferingID  types.String `tfsdk:"offering_id"`
	ID          types.String `tfsdk:"id"`
	LookupKey   types.String `tfsdk:"lookup_key"`
	DisplayName types.String `tfsdk:"display_name"`
	Position    types.Int64  `tfsdk:"position"`
	CreatedAt   types.Int64  `tfsdk:"created_at"`
}

type packageResponse struct {
	ID          string `json:"id"`
	LookupKey   string `json:"lookup_key"`
	DisplayName string `json:"display_name"`
	Position    int64  `json:"position"`
	CreatedAt   int64  `json:"created_at"`
}

func NewPackageResource() resource.Resource { return &PackageResource{} }

func (r *PackageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_package"
}

func (r *PackageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "A package within a RevenueCat offering.", Attributes: map[string]schema.Attribute{
		"project_id":   schema.StringAttribute{Required: true, MarkdownDescription: "RevenueCat project ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"offering_id":  schema.StringAttribute{Required: true, MarkdownDescription: "Parent offering ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"id":           schema.StringAttribute{Computed: true, MarkdownDescription: "RevenueCat package ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"lookup_key":   schema.StringAttribute{Required: true, MarkdownDescription: "Stable package lookup key.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"display_name": schema.StringAttribute{Required: true, MarkdownDescription: "Human-readable package name."},
		"position":     schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(1), MarkdownDescription: "Package position within the offering."},
		"created_at":   schema.Int64Attribute{Computed: true, MarkdownDescription: "Creation time in milliseconds since Unix epoch."},
	}}
}

func (r *PackageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureClient(req, resp, &r.client)
}

func (r *PackageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PackageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var remote packageResponse
	err := r.client.Do(ctx, http.MethodPost, escaped("projects", data.ProjectID.ValueString(), "offerings", data.OfferingID.ValueString(), "packages"), map[string]any{"lookup_key": data.LookupKey.ValueString(), "display_name": data.DisplayName.ValueString(), "position": data.Position.ValueInt64()}, &remote)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create RevenueCat package", err.Error())
		return
	}
	setPackageState(&data, remote)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PackageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PackageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var remote packageResponse
	err := r.client.Do(ctx, http.MethodGet, escaped("projects", data.ProjectID.ValueString(), "packages", data.ID.ValueString()), nil, &remote)
	if readError(err, ctx, resp, "package") {
		return
	}
	setPackageState(&data, remote)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PackageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PackageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var remote packageResponse
	err := r.client.Do(ctx, http.MethodPost, escaped("projects", data.ProjectID.ValueString(), "packages", data.ID.ValueString()), map[string]any{"display_name": data.DisplayName.ValueString(), "position": data.Position.ValueInt64()}, &remote)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update RevenueCat package", err.Error())
		return
	}
	setPackageState(&data, remote)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PackageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PackageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Do(ctx, http.MethodDelete, escaped("projects", data.ProjectID.ValueString(), "packages", data.ID.ValueString()), nil, nil)
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to delete RevenueCat package", err.Error())
	}
}

func (r *PackageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importThreePartID(ctx, req, resp, "offering_id")
}

func setPackageState(data *PackageResourceModel, remote packageResponse) {
	data.ID = types.StringValue(remote.ID)
	data.LookupKey = types.StringValue(remote.LookupKey)
	data.DisplayName = types.StringValue(remote.DisplayName)
	data.Position = types.Int64Value(remote.Position)
	data.CreatedAt = types.Int64Value(remote.CreatedAt)
}
