# Publishing the provider

The Terraform Registry and OpenTofu Registry both install provider binaries from
signed GitHub releases. The same release artifacts can serve both registries,
but each registry requires a separate one-time registration.

## Repository readiness

Before publishing, confirm:

- the public repository is named `terraform-provider-revenuecat`;
- generated provider and resource documentation is committed under `docs/`;
- `terraform-registry-manifest.json` declares protocol version `6.0`;
- the release workflow creates platform ZIP archives, a manifest, SHA-256
  checksums, and a detached signature; and
- the default branch contains the exact commit being released.

Do not reuse or replace an existing release version. Publish a new semantic
version whenever release artifacts change.

## One-time signing setup

HashiCorp requires signed provider releases. Create a dedicated RSA GPG key;
the Terraform Registry does not accept the default ECC key type. Protect and
back up both the key and its passphrase.

Export the keys in ASCII-armored form:

```shell
gpg --armor --export "KEY_ID_OR_EMAIL" > provider-public-key.asc
gpg --armor --export-secret-keys "KEY_ID_OR_EMAIL" > provider-private-key.asc
```

Store the contents of `provider-private-key.asc` as the GitHub Actions secret
`GPG_PRIVATE_KEY`, and store the key passphrase as `PASSPHRASE`. Never commit
either file. Register the public key separately with each registry as described
below.

## Create the first GitHub release

After merging the release-ready pull request to the default branch:

```shell
git switch main
git pull --ff-only
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

The tag starts `.github/workflows/release.yml`. Verify that the published GitHub
release contains ZIP files for the supported platforms plus these files:

```text
terraform-provider-revenuecat_0.1.0_manifest.json
terraform-provider-revenuecat_0.1.0_SHA256SUMS
terraform-provider-revenuecat_0.1.0_SHA256SUMS.sig
```

The release must be public and finalized, not a draft. Do not create a branch
named `v0.1.0`.

## Terraform Registry registration

1. Sign in to the [Terraform Registry](https://registry.terraform.io/) with the
   GitHub account that can administer the `cloudridgeworks` organization.
2. Authorize the Terraform Registry GitHub application for the public provider
   repository.
3. In **User Settings > Signing Keys**, add the ASCII-armored public RSA key for
   the `cloudridgeworks` namespace.
4. Choose **Publish > Provider**, select `cloudridgeworks` and
   `terraform-provider-revenuecat`, and complete the prompts.
5. Confirm that `cloudridgeworks/revenuecat` resolves and that a clean
   `terraform init` verifies the signed `v0.1.0` release.

The Registry creates a GitHub webhook for future release events. Later versions
only require a new `vX.Y.Z` tag and successful release workflow. Use the
provider's Registry **Resync** action if a release is not discovered.

Official instructions:
[Publish providers](https://developer.hashicorp.com/terraform/registry/providers/publishing).

## OpenTofu Registry registration

OpenTofu registration must be performed through the OpenTofu Registry's GitHub
issue form UI. Its automation does not process pull requests or issues created
through the GitHub CLI or API.

1. Make the submitting account's membership in the `cloudridgeworks` GitHub
   organization public.
2. Use **Submit new Provider** in the
   [OpenTofu Registry repository](https://github.com/opentofu/registry/issues/new/choose)
   and enter `cloudridgeworks/terraform-provider-revenuecat`.
3. After the provider submission is accepted, use **Submit new Provider Signing
   Key** in the same issue-form UI. Enter namespace `cloudridgeworks`, provider
   name `revenuecat`, and paste the ASCII-armored public GPG key.
4. Confirm that `cloudridgeworks/revenuecat` appears in the
   [OpenTofu Registry](https://search.opentofu.org/) and that a clean `tofu init`
   installs and verifies `v0.1.0`.

Registration is required once. The OpenTofu Registry automatically discovers
later signed GitHub releases.

Official instructions:
[Adding a provider](https://search.opentofu.org/docs/providers/adding) and the
[OpenTofu Registry submission policy](https://github.com/opentofu/registry/blob/main/README.md#adding-providers-modules-or-gpg-keys-to-the-opentofu-registry).

## Subsequent releases

1. Update `CHANGELOG.md` and generated documentation.
2. Merge a fully tested release commit to `main`.
3. Create and push a new immutable `vX.Y.Z` tag.
4. Verify the release workflow, signature, checksums, and both registry pages.
5. Test installation with fresh Terraform and OpenTofu working directories.
