resource "revenuecat_offering" "default" {
  project_id   = var.revenuecat_project_id
  lookup_key   = "default"
  display_name = "PresencePath Pro"
  is_current   = true
}
