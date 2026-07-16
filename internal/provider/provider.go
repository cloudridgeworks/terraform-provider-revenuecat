// Copyright Cloud Ridge Works
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudridgeworks/terraform-provider-revenuecat/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &RevenueCatProvider{}

type RevenueCatProvider struct{ version string }

type RevenueCatProviderModel struct {
	APIKey  types.String `tfsdk:"api_key"`
	BaseURL types.String `tfsdk:"base_url"`
}

func (p *RevenueCatProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "revenuecat"
	resp.Version = p.version
}

func (p *RevenueCatProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage RevenueCat project configuration through the RevenueCat REST API v2.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "RevenueCat v2 secret API key. May also be set with `REVENUECAT_API_KEY`.",
				Optional:            true,
				Sensitive:           true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "RevenueCat REST API v2 base URL. May also be set with `REVENUECAT_BASE_URL`; intended for testing.",
				Optional:            true,
			},
		},
	}
}

func (p *RevenueCatProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data RevenueCatProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.APIKey.IsUnknown() || data.BaseURL.IsUnknown() {
		resp.Diagnostics.AddError("Unknown RevenueCat provider configuration", "api_key and base_url must be known before provider configuration")
		return
	}

	apiKey := os.Getenv("REVENUECAT_API_KEY")
	if !data.APIKey.IsNull() {
		apiKey = data.APIKey.ValueString()
	}
	baseURL := os.Getenv("REVENUECAT_BASE_URL")
	if !data.BaseURL.IsNull() {
		baseURL = data.BaseURL.ValueString()
	}
	c, err := client.New(apiKey, baseURL, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to configure RevenueCat client", err.Error())
		return
	}
	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *RevenueCatProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewEntitlementResource,
		NewOfferingResource,
		NewPackageResource,
		NewEntitlementProductResource,
		NewPackageProductResource,
		NewWebhookResource,
	}
}

func (p *RevenueCatProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider { return &RevenueCatProvider{version: version} }
}

func configureClient(req resource.ConfigureRequest, resp *resource.ConfigureResponse, target **client.Client) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	*target = c
}
