// Copyright Cloud Ridge Works
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEntitlementResource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC=1 to run acceptance tests")
	}
	var mu sync.Mutex
	remote := entitlementResponse{ID: "ent_test", LookupKey: "pro", DisplayName: "PresencePath Pro", State: "active", CreatedAt: 1710000000000}
	deleted := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if req.Header.Get("Authorization") != "Bearer test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		collectionPath := "/v2/projects/proj_test/entitlements"
		resourcePath := collectionPath + "/ent_test"
		switch {
		case req.Method == http.MethodPost && req.URL.Path == collectionPath:
			var body struct {
				LookupKey   string `json:"lookup_key"`
				DisplayName string `json:"display_name"`
			}
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			remote.LookupKey, remote.DisplayName, deleted = body.LookupKey, body.DisplayName, false
			writeJSON(w, http.StatusCreated, remote)
		case req.Method == http.MethodGet && req.URL.Path == resourcePath:
			if deleted {
				http.NotFound(w, req)
				return
			}
			writeJSON(w, http.StatusOK, remote)
		case req.Method == http.MethodPost && req.URL.Path == resourcePath:
			var body struct {
				DisplayName string `json:"display_name"`
			}
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			remote.DisplayName = body.DisplayName
			writeJSON(w, http.StatusOK, remote)
		case req.Method == http.MethodDelete && req.URL.Path == resourcePath:
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "unexpected request", http.StatusNotFound)
		}
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: entitlementTestConfig(server.URL, "PresencePath Pro"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("revenuecat_entitlement.test", "id", "ent_test"),
					resource.TestCheckResourceAttr("revenuecat_entitlement.test", "state", "active"),
				),
			},
			{
				Config: entitlementTestConfig(server.URL, "PresencePath Premium"),
				Check:  resource.TestCheckResourceAttr("revenuecat_entitlement.test", "display_name", "PresencePath Premium"),
			},
			{
				ResourceName:      "revenuecat_entitlement.test",
				ImportState:       true,
				ImportStateId:     "proj_test/ent_test",
				ImportStateVerify: true,
			},
		},
	})
}

func entitlementTestConfig(baseURL, displayName string) string {
	return fmt.Sprintf(`
provider "revenuecat" {
  api_key = "test-key"
  base_url = %q
}

resource "revenuecat_entitlement" "test" {
  project_id = "proj_test"
  lookup_key = "pro"
  display_name = %q
}
`, baseURL+"/v2", displayName)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
