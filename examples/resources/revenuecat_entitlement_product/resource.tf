# RevenueCat API v2 cannot create Web Billing products. Create the product in
# RevenueCat first, then attach its RevenueCat product ID through this resource.
resource "revenuecat_entitlement_product" "web_monthly" {
  project_id     = var.revenuecat_project_id
  entitlement_id = revenuecat_entitlement.pro.id
  product_id     = var.revenuecat_web_monthly_product_id
}
