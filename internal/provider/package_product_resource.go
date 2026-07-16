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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &PackageProductResource{}
var _ resource.ResourceWithConfigure = &PackageProductResource{}
var _ resource.ResourceWithImportState = &PackageProductResource{}

type PackageProductResource struct{ client *client.Client }

type PackageProductResourceModel struct {
	ProjectID           types.String `tfsdk:"project_id"`
	PackageID           types.String `tfsdk:"package_id"`
	ProductID           types.String `tfsdk:"product_id"`
	EligibilityCriteria types.String `tfsdk:"eligibility_criteria"`
	ID                  types.String `tfsdk:"id"`
}

func NewPackageProductResource() resource.Resource { return &PackageProductResource{} }

func (r *PackageProductResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_package_product"
}

func (r *PackageProductResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{MarkdownDescription: "Attaches an existing RevenueCat product to a package.", Attributes: map[string]schema.Attribute{
		"project_id":           schema.StringAttribute{Required: true, MarkdownDescription: "RevenueCat project ID.", PlanModifiers: replace},
		"package_id":           schema.StringAttribute{Required: true, MarkdownDescription: "RevenueCat package ID.", PlanModifiers: replace},
		"product_id":           schema.StringAttribute{Required: true, MarkdownDescription: "Existing RevenueCat product ID.", PlanModifiers: replace},
		"eligibility_criteria": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("all"), MarkdownDescription: "One of `all`, `google_sdk_lt_6`, or `google_sdk_ge_6`.", PlanModifiers: replace},
		"id":                   schema.StringAttribute{Computed: true, MarkdownDescription: "Stable association identifier.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}

func (r *PackageProductResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureClient(req, resp, &r.client)
}

func (r *PackageProductResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PackageProductResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	endpoint := escaped("projects", data.ProjectID.ValueString(), "packages", data.PackageID.ValueString(), "actions", "attach_products")
	body := map[string]any{"products": []map[string]string{{"product_id": data.ProductID.ValueString(), "eligibility_criteria": data.EligibilityCriteria.ValueString()}}}
	if err := r.client.Do(ctx, http.MethodPost, endpoint, body, nil); err != nil {
		resp.Diagnostics.AddError("Unable to attach RevenueCat product to package", err.Error())
		return
	}
	data.ID = types.StringValue(data.PackageID.ValueString() + "/" + data.ProductID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PackageProductResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PackageProductResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	endpoint := escaped("projects", data.ProjectID.ValueString(), "packages", data.PackageID.ValueString(), "products")
	criteria, found, err := r.client.FindProductAssociation(ctx, endpoint, data.ProductID.ValueString())
	if errors.Is(err, client.ErrNotFound) || (err == nil && !found) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read RevenueCat package product", err.Error())
		return
	}
	data.EligibilityCriteria = types.StringValue(criteria)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PackageProductResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PackageProductResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	}
}

func (r *PackageProductResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PackageProductResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	endpoint := escaped("projects", data.ProjectID.ValueString(), "packages", data.PackageID.ValueString(), "actions", "detach_products")
	if err := r.client.Do(ctx, http.MethodPost, endpoint, map[string]any{"product_ids": []string{data.ProductID.ValueString()}}, nil); err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to detach RevenueCat product from package", err.Error())
	}
}

func (r *PackageProductResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import identifier", "Expected project_id/package_id/product_id")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("package_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("product_id"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1]+"/"+parts[2])...)
}
