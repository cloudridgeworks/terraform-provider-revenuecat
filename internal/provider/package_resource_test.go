// Copyright Cloud Ridge Works 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAddConfiguredPosition(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		position types.Int64
		want     any
		present  bool
	}{
		"configured": {position: types.Int64Value(2), want: int64(2), present: true},
		"null":       {position: types.Int64Null()},
		"unknown":    {position: types.Int64Unknown()},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			body := map[string]any{"display_name": "Monthly"}
			addConfiguredPosition(body, test.position)
			got, present := body["position"]
			if present != test.present {
				t.Fatalf("position presence = %t, want %t", present, test.present)
			}
			if present && got != test.want {
				t.Fatalf("position = %#v, want %#v", got, test.want)
			}
		})
	}
}
