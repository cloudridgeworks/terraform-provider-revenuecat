resource "revenuecat_webhook" "subscription_events" {
  project_id           = var.revenuecat_project_id
  name                 = "PresencePath subscription events"
  url                  = "https://api.example.com/v1/webhooks/revenuecat"
  authorization_header = var.revenuecat_webhook_authorization_header
  environment          = "production"
}
