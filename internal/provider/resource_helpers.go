// Copyright Cloud Ridge Works
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/cloudridgeworks/terraform-provider-revenuecat/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

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
