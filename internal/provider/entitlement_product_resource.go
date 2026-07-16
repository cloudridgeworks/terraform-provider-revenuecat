// Copyright Cloud Ridge Works
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/cloudridgeworks/terraform-provider-revenuecat/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &EntitlementProductResource{}
var _ resource.ResourceWithConfigure = &EntitlementProductResource{}
var _ resource.ResourceWithImportState = &EntitlementProductResource{}

type EntitlementProductResource struct{ client *client.Client }

type EntitlementProductResourceModel struct {
	ProjectID     types.String `tfsdk:"project_id"`
	EntitlementID types.String `tfsdk:"entitlement_id"`
	ProductID     types.String `tfsdk:"product_id"`
	ID            types.String `tfsdk:"id"`
}

func NewEntitlementProductResource() resource.Resource { return &EntitlementProductResource{} }

func (r *EntitlementProductResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entitlement_product"
}

func (r *EntitlementProductResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{MarkdownDescription: "Attaches an existing RevenueCat product to an entitlement.", Attributes: map[string]schema.Attribute{
		"project_id":     schema.StringAttribute{Required: true, MarkdownDescription: "RevenueCat project ID.", PlanModifiers: replace},
		"entitlement_id": schema.StringAttribute{Required: true, MarkdownDescription: "RevenueCat entitlement ID.", PlanModifiers: replace},
		"product_id":     schema.StringAttribute{Required: true, MarkdownDescription: "Existing RevenueCat product ID.", PlanModifiers: replace},
		"id":             schema.StringAttribute{Computed: true, MarkdownDescription: "Stable association identifier.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}

func (r *EntitlementProductResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureClient(req, resp, &r.client)
}

func (r *EntitlementProductResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EntitlementProductResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	endpoint := escaped("projects", data.ProjectID.ValueString(), "entitlements", data.EntitlementID.ValueString(), "actions", "attach_products")
	err := r.client.Do(ctx, http.MethodPost, endpoint, map[string]any{"product_ids": []string{data.ProductID.ValueString()}}, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to attach RevenueCat product to entitlement", err.Error())
		return
	}
	data.ID = types.StringValue(data.EntitlementID.ValueString() + "/" + data.ProductID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EntitlementProductResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EntitlementProductResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	endpoint := escaped("projects", data.ProjectID.ValueString(), "entitlements", data.EntitlementID.ValueString(), "products")
	_, found, err := r.client.FindProductAssociation(ctx, endpoint, data.ProductID.ValueString())
	if errors.Is(err, client.ErrNotFound) || (err == nil && !found) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read RevenueCat entitlement product", err.Error())
	}
}

func (r *EntitlementProductResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EntitlementProductResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	}
}

func (r *EntitlementProductResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EntitlementProductResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	endpoint := escaped("projects", data.ProjectID.ValueString(), "entitlements", data.EntitlementID.ValueString(), "actions", "detach_products")
	err := r.client.Do(ctx, http.MethodPost, endpoint, map[string]any{"product_ids": []string{data.ProductID.ValueString()}}, nil)
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to detach RevenueCat product from entitlement", err.Error())
	}
}

func (r *EntitlementProductResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import identifier", "Expected project_id/entitlement_id/product_id")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("entitlement_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("product_id"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1]+"/"+parts[2])...)
}
