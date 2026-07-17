# Terraform Provider for RevenueCat

An experimental Terraform and OpenTofu provider for managing RevenueCat project
configuration through the [RevenueCat REST API v2](https://www.revenuecat.com/docs/api-v2).
It follows HashiCorp's
[`terraform-provider-scaffolding-framework`](https://github.com/hashicorp/terraform-provider-scaffolding-framework)
layout and uses Terraform Plugin Framework protocol 6.

## Resource support

| RevenueCat resource or capability | Status | Terraform resource | Reason or limitation |
| --- | --- | --- | --- |
| Entitlements | Supported | `revenuecat_entitlement` | Create, read, update, delete, import, and drift detection are implemented. |
| Offerings | Supported | `revenuecat_offering` | Create, read, update, delete, import, current-offering selection, and drift detection are implemented; offering metadata is not managed. |
| Packages | Supported | `revenuecat_package` | Create, read, update, delete, import, ordering, and drift detection are implemented. |
| Entitlement-product attachments | Supported | `revenuecat_entitlement_product` | One product association is managed per resource; association changes replace the resource. |
| Package-product attachments | Supported | `revenuecat_package_product` | One product association and its eligibility criteria are managed per resource; association changes replace the resource. |
| Webhook integrations | Supported | `revenuecat_webhook` | Create, read, update, delete, import, filters, app scoping, and drift detection are implemented. |
| Web Billing products | Unsupported upstream | None | RevenueCat API v2 explicitly does not permit creating Web Billing products. Create them externally, then attach their RevenueCat product IDs with the supported attachment resources. |
| Native-store and Stripe-backed products | Not implemented | None | RevenueCat API v2 supports some product creation, but product lifecycle management is outside the initial provider surface. Existing products can be attached by ID. |
| Offering metadata | Not implemented | None | RevenueCat API v2 supports offering metadata, but the initial offering resource does not expose it. |
| Projects and apps | Not implemented | None | The provider expects an existing project and accepts `project_id`; project and app bootstrapping remain external. |
| Paywalls and other integrations | Not implemented | None | No provider resources are currently implemented for these project-configuration objects, and RevenueCat API coverage varies by object. |
| Customers, subscriptions, and transactions | Out of scope | None | These are operational/customer records rather than declarative project configuration and are intentionally not managed as Terraform resources. |

All supported resources refresh remote state, remove externally deleted objects from
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

## OpenTofu compatibility

The provider uses Terraform Plugin Framework protocol 6, which is supported by
OpenTofu. After the provider is added to the OpenTofu Registry, the same source
address and configuration work without code changes:

```shell
tofu init
tofu plan
tofu apply
```

Publishing to the Terraform Registry does not automatically publish a new
provider to the separate OpenTofu Registry. The provider repository must be
submitted to the OpenTofu Registry once; later GitHub releases are discovered
automatically.

## Development

```shell
make test
make build
make generate
```

Generated provider documentation lives in [`docs`](docs). Releases use the
scaffold's GoReleaser workflow and require a configured GPG signing key before
publishing to the Terraform Registry.

Please read [`CONTRIBUTING.md`](CONTRIBUTING.md) before opening an issue or pull
request. Security vulnerabilities should be reported privately as described in
[`SECURITY.md`](SECURITY.md).

## Publishing

The repository is prepared to publish the same signed GitHub release artifacts
to both registries. See [`PUBLISHING.md`](PUBLISHING.md) for the one-time GPG and
registry setup, first-release checklist, and subsequent release process.

## License

This project is licensed under the [Mozilla Public License 2.0](LICENSE).
Contributions are accepted under the same license; no contributor license
agreement is currently required. See [`NOTICE`](NOTICE) for attribution and
trademark information.
