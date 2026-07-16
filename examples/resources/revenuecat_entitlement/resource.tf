resource "revenuecat_entitlement" "pro" {
  project_id   = var.revenuecat_project_id
  lookup_key   = "pro"
  display_name = "PresencePath Pro"
}
