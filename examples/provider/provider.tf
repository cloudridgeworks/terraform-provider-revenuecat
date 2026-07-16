terraform {
  required_providers {
    revenuecat = {
      source  = "cloudridgeworks/revenuecat"
      version = "~> 0.1"
    }
  }
}

# Set REVENUECAT_API_KEY to a RevenueCat v2 secret key with the project
# configuration permissions needed by the resources in this configuration.
provider "revenuecat" {}
