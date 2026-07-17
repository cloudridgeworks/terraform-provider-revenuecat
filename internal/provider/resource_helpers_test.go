// Copyright Cloud Ridge Works 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAbsoluteURIValidator(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value   types.String
		wantErr bool
	}{
		"https URI": {value: types.StringValue("https://example.com/revenuecat")},
		"http URI":  {value: types.StringValue("http://localhost:8080/webhook")},
		"relative":  {value: types.StringValue("/webhook"), wantErr: true},
		"no host":   {value: types.StringValue("https:///webhook"), wantErr: true},
		"invalid":   {value: types.StringValue(":not-a-uri"), wantErr: true},
		"null":      {value: types.StringNull()},
		"unknown":   {value: types.StringUnknown()},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var resp validator.StringResponse
			absoluteURIValidator{}.ValidateString(context.Background(), validator.StringRequest{
				Path:        path.Root("url"),
				ConfigValue: test.value,
			}, &resp)
			if got := resp.Diagnostics.HasError(); got != test.wantErr {
				t.Fatalf("HasError() = %t, want %t; diagnostics: %v", got, test.wantErr, resp.Diagnostics)
			}
		})
	}
}
