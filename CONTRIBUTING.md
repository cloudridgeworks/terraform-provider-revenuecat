# Contributing

Thanks for helping improve the RevenueCat provider. Bug reports, documentation
fixes, tests, and focused provider enhancements are welcome.

## Before opening an issue

- Search existing issues and confirm the behavior against RevenueCat API v2.
- Use the bug report form for defects and the feature request form for proposed
  resources or behavior.
- Remove API keys, project IDs, signing secrets, customer data, and other
  sensitive values from logs and examples.
- Report suspected vulnerabilities privately according to
  [`SECURITY.md`](SECURITY.md), not in a public issue.

## Development setup

Development requires Go 1.25.8 or newer, Terraform 1.0 or a compatible OpenTofu
release, and GNU Make.

```shell
go mod download
make test
make build
make generate
```

The acceptance suite uses a local mock RevenueCat server and does not require a
real API key:

```shell
make testacc
```

Live testing must use a disposable or explicitly approved RevenueCat project.
Never commit a RevenueCat secret key or put it in Terraform configuration,
state, test output, or fixtures.

## Pull requests

1. Keep each pull request focused and explain the user-visible behavior.
2. Add tests for new behavior, including import and drift handling where
   applicable.
3. Update examples and schema descriptions when configuration changes.
4. Run `make generate` and commit generated files under `docs/`.
5. Run `make test`, `make testacc`, `make build`, and the configured linter.
6. Call out breaking changes and RevenueCat API limitations explicitly.

Generated documentation should not be edited directly. Change the provider
schema, templates, or examples and regenerate it instead.

## Commit and review expectations

Use clear, imperative commit messages. Maintainers may request a changelog entry,
additional platform testing, or a live RevenueCat smoke test before merging a
change that affects API behavior. Pull requests are squash-merged after required
checks and review pass.

## Licensing contributions

The project is licensed under the Mozilla Public License 2.0. By submitting a
contribution, you represent that you have the right to submit it and agree that
it is provided under the terms of the project [`LICENSE`](LICENSE). No separate
contributor license agreement is currently required.

This project is independently maintained and is not affiliated with or endorsed
by RevenueCat, HashiCorp, or the OpenTofu project.
