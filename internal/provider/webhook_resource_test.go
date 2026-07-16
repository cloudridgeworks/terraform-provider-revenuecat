// Copyright Cloud Ridge Works
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSetWebhookStateHandlesOmittedSigningSecret(t *testing.T) {
	data := WebhookResourceModel{SigningSecret: types.StringUnknown()}
	remote := webhookResponse{
		ID:        "wh_test",
		Name:      "Test webhook",
		URL:       "https://example.com/webhook",
		CreatedAt: 1710000000000,
	}

	diags := setWebhookState(context.Background(), &data, remote)
	if diags.HasError() {
		t.Fatalf("setWebhookState returned diagnostics: %v", diags)
	}
	if !data.SigningSecret.IsNull() {
		t.Fatalf("expected omitted signing_secret to become null, got %#v", data.SigningSecret)
	}
}

func TestSetWebhookStatePreservesKnownSigningSecret(t *testing.T) {
	data := WebhookResourceModel{SigningSecret: types.StringValue("existing-secret")}
	remote := webhookResponse{
		ID:        "wh_test",
		Name:      "Test webhook",
		URL:       "https://example.com/webhook",
		CreatedAt: 1710000000000,
	}

	diags := setWebhookState(context.Background(), &data, remote)
	if diags.HasError() {
		t.Fatalf("setWebhookState returned diagnostics: %v", diags)
	}
	if got := data.SigningSecret.ValueString(); got != "existing-secret" {
		t.Fatalf("expected known signing_secret to be preserved, got %q", got)
	}
}
