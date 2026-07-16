# Terraform Provider for RevenueCat

An experimental Terraform and OpenTofu provider for managing RevenueCat project
configuration through the [RevenueCat REST API v2](https://www.revenuecat.com/docs/api-v2).
It follows HashiCorp's
[`terraform-provider-scaffolding-framework`](https://github.com/hashicorp/terraform-provider-scaffolding-framework)
layout and uses Terraform Plugin Framework protocol 6.

## Supported resources

- `revenuecat_entitlement`
- `revenuecat_offering`
- `revenuecat_package`
- `revenuecat_entitlement_product`
- `revenuecat_package_product`
- `revenuecat_webhook`

All resources refresh remote state, remove externally deleted objects from
state, and support import. Product attachment resources intentionally manage
one relationship each so separate modules do not fight over an entire product
set.

## API limitation

RevenueCat API v2 does not allow creating Web Billing products. Create those
products in RevenueCat first, then pass their RevenueCat product IDs to the two
product attachment resources. Native or Stripe product creation is not included
in the initial provider surface.

## Requirements

- Terraform >= 1.0 or a compatible OpenTofu release
- Go >= 1.25.8 for provider development
- A RevenueCat v2 secret key with the required `project_configuration` read and
  read/write permissions

## Usage

```hcl
terraform {
  required_providers {
    revenuecat = {
      source  = "cloudridgeworks/revenuecat"
      version = "~> 0.1"
    }
  }
}

provider "revenuecat" {}

resource "revenuecat_entitlement" "pro" {
  project_id   = var.revenuecat_project_id
  lookup_key   = "pro"
  display_name = "PresencePath Pro"
}
```

Set the secret key outside configuration:

```shell
export REVENUECAT_API_KEY="your-v2-secret-key"
```

See [`examples/resources`](examples/resources) for complete resource examples.

## Development

```shell
make test
make build
make generate
```

Generated provider documentation lives in [`docs`](docs). Releases use the
scaffold's GoReleaser workflow and require a configured GPG signing key before
publishing to the Terraform Registry.

## License

Mozilla Public License 2.0. See [`LICENSE`](LICENSE).
