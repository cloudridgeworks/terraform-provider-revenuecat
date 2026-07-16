// Copyright Cloud Ridge Works
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"net/http"

	"github.com/cloudridgeworks/terraform-provider-revenuecat/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &WebhookResource{}
var _ resource.ResourceWithConfigure = &WebhookResource{}
var _ resource.ResourceWithImportState = &WebhookResource{}

type WebhookResource struct{ client *client.Client }

type WebhookResourceModel struct {
	ProjectID           types.String `tfsdk:"project_id"`
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	URL                 types.String `tfsdk:"url"`
	AuthorizationHeader types.String `tfsdk:"authorization_header"`
	Environment         types.String `tfsdk:"environment"`
	EventTypes          types.Set    `tfsdk:"event_types"`
	AppID               types.String `tfsdk:"app_id"`
	SigningSecret       types.String `tfsdk:"signing_secret"`
	CreatedAt           types.Int64  `tfsdk:"created_at"`
}

type webhookResponse struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	URL           string   `json:"url"`
	Environment   *string  `json:"environment"`
	EventTypes    []string `json:"event_types"`
	AppID         *string  `json:"app_id"`
	SigningSecret string   `json:"signing_secret"`
	CreatedAt     int64    `json:"created_at"`
}

func NewWebhookResource() resource.Resource { return &WebhookResource{} }

func (r *WebhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *WebhookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "A RevenueCat webhook integration.", Attributes: map[string]schema.Attribute{
		"project_id":           schema.StringAttribute{Required: true, MarkdownDescription: "RevenueCat project ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"id":                   schema.StringAttribute{Computed: true, MarkdownDescription: "RevenueCat webhook integration ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"name":                 schema.StringAttribute{Required: true, MarkdownDescription: "Webhook display name."},
		"url":                  schema.StringAttribute{Required: true, MarkdownDescription: "HTTPS endpoint that receives RevenueCat events."},
		"authorization_header": schema.StringAttribute{Optional: true, Sensitive: true, MarkdownDescription: "Optional Authorization header sent with webhook requests."},
		"environment":          schema.StringAttribute{Optional: true, MarkdownDescription: "Optional `production` or `sandbox` event filter."},
		"event_types":          schema.SetAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Optional set of RevenueCat webhook event types."},
		"app_id":               schema.StringAttribute{Optional: true, MarkdownDescription: "Optional RevenueCat app ID scope."},
		"signing_secret":       schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "RevenueCat webhook signing secret."},
		"created_at":           schema.Int64Attribute{Computed: true, MarkdownDescription: "Creation time in milliseconds since Unix epoch."},
	}}
}

func (r *WebhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureClient(req, resp, &r.client)
}

func (r *WebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WebhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, diags := webhookBody(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var remote webhookResponse
	if err := r.client.Do(ctx, http.MethodPost, escaped("projects", data.ProjectID.ValueString(), "integrations", "webhooks"), body, &remote); err != nil {
		resp.Diagnostics.AddError("Unable to create RevenueCat webhook", err.Error())
		return
	}
	resp.Diagnostics.Append(setWebhookState(ctx, &data, remote)...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	}
}

func (r *WebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WebhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var remote webhookResponse
	err := r.client.Do(ctx, http.MethodGet, escaped("projects", data.ProjectID.ValueString(), "integrations", "webhooks", data.ID.ValueString()), nil, &remote)
	if readError(err, ctx, resp, "webhook") {
		return
	}
	resp.Diagnostics.Append(setWebhookState(ctx, &data, remote)...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	}
}

func (r *WebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WebhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, diags := webhookBody(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var remote webhookResponse
	if err := r.client.Do(ctx, http.MethodPost, escaped("projects", data.ProjectID.ValueString(), "integrations", "webhooks", data.ID.ValueString()), body, &remote); err != nil {
		resp.Diagnostics.AddError("Unable to update RevenueCat webhook", err.Error())
		return
	}
	resp.Diagnostics.Append(setWebhookState(ctx, &data, remote)...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	}
}

func (r *WebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WebhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.Do(ctx, http.MethodDelete, escaped("projects", data.ProjectID.ValueString(), "integrations", "webhooks", data.ID.ValueString()), nil, nil)
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to delete RevenueCat webhook", err.Error())
	}
}

func (r *WebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importTwoPartID(ctx, req, resp)
}

func webhookBody(ctx context.Context, data WebhookResourceModel) (map[string]any, diag.Diagnostics) {
	var eventTypes []string
	var diags diag.Diagnostics
	var events any
	if !data.EventTypes.IsNull() {
		diags = data.EventTypes.ElementsAs(ctx, &eventTypes, false)
		events = eventTypes
	}
	return map[string]any{
		"name": data.Name.ValueString(), "url": data.URL.ValueString(),
		"authorization_header": nullableString(data.AuthorizationHeader),
		"environment":          nullableString(data.Environment), "event_types": events,
		"app_id": nullableString(data.AppID),
	}, diags
}

func nullableString(value types.String) any {
	if value.IsNull() {
		return nil
	}
	return value.ValueString()
}

func setWebhookState(ctx context.Context, data *WebhookResourceModel, remote webhookResponse) diag.Diagnostics {
	data.ID = types.StringValue(remote.ID)
	data.Name = types.StringValue(remote.Name)
	data.URL = types.StringValue(remote.URL)
	if remote.Environment == nil {
		data.Environment = types.StringNull()
	} else {
		data.Environment = types.StringValue(*remote.Environment)
	}
	if remote.AppID == nil {
		data.AppID = types.StringNull()
	} else {
		data.AppID = types.StringValue(*remote.AppID)
	}
	if remote.EventTypes == nil {
		data.EventTypes = types.SetNull(types.StringType)
	} else {
		set, diags := types.SetValueFrom(ctx, types.StringType, remote.EventTypes)
		if diags.HasError() {
			return diags
		}
		data.EventTypes = set
	}
	if remote.SigningSecret != "" {
		data.SigningSecret = types.StringValue(remote.SigningSecret)
	} else if data.SigningSecret.IsUnknown() {
		data.SigningSecret = types.StringNull()
	}
	data.CreatedAt = types.Int64Value(remote.CreatedAt)
	return nil
}
