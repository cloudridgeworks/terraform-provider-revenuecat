// Copyright Cloud Ridge Works 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/cloudridgeworks/terraform-provider-revenuecat/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	maxIdentifierLength  = 255
	maxLookupKeyLength   = 200
	maxDisplayNameLength = 1500
	maxWebhookURLLength  = 5000
)

func identifierValidators() []validator.String {
	return []validator.String{stringvalidator.UTF8LengthBetween(1, maxIdentifierLength)}
}

func lookupKeyValidators() []validator.String {
	return []validator.String{stringvalidator.UTF8LengthBetween(1, maxLookupKeyLength)}
}

func displayNameValidators() []validator.String {
	return []validator.String{stringvalidator.UTF8LengthBetween(1, maxDisplayNameLength)}
}

func webhookURLValidators() []validator.String {
	return []validator.String{
		stringvalidator.UTF8LengthBetween(1, maxWebhookURLLength),
		absoluteURIValidator{},
	}
}

var _ validator.String = absoluteURIValidator{}

type absoluteURIValidator struct{}

func (absoluteURIValidator) Description(_ context.Context) string {
	return "value must be a valid absolute URI"
}

func (v absoluteURIValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v absoluteURIValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	parsed, err := url.ParseRequestURI(req.ConfigValue.ValueString())
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid URI", v.Description(ctx))
	}
}

func escaped(parts ...string) string {
	encoded := make([]string, len(parts))
	for i, part := range parts {
		encoded[i] = url.PathEscape(part)
	}
	return strings.Join(encoded, "/")
}

func readError(err error, ctx context.Context, resp *resource.ReadResponse, noun string) bool {
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return true
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read RevenueCat "+noun, err.Error())
		return true
	}
	return false
}

func importTwoPartID(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import identifier", "Expected project_id/resource_id")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func importThreePartID(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse, middleAttribute string) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError("Invalid import identifier", fmt.Sprintf("Expected project_id/%s/resource_id", middleAttribute))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(middleAttribute), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[2])...)
}
