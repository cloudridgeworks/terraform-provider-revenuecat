resource "revenuecat_package_product" "web_monthly" {
  project_id           = var.revenuecat_project_id
  package_id           = revenuecat_package.monthly.id
  product_id           = var.revenuecat_web_monthly_product_id
  eligibility_criteria = "all"
}
