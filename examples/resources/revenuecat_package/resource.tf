resource "revenuecat_package" "monthly" {
  project_id   = var.revenuecat_project_id
  offering_id  = revenuecat_offering.default.id
  lookup_key   = "$rc_monthly"
  display_name = "Monthly"
  position     = 1
}
