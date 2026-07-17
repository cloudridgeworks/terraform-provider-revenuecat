# Security policy

## Supported versions

Until the first stable release, security fixes are made on the latest release
line only. After `v1.0.0`, this policy will be updated with the supported release
lines.

## Reporting a vulnerability

Do not open a public issue for a suspected vulnerability or include secrets in
an issue, discussion, pull request, or test log.

Use GitHub's **Report a vulnerability** action on the repository's Security tab
to open a private security advisory. Include:

- the affected provider version and platform;
- reproduction steps or a minimal configuration with all secrets removed;
- the expected and observed impact; and
- any suggested mitigation.

If private vulnerability reporting is unavailable, contact a repository
maintainer privately through their GitHub profile before sharing technical
details. Maintainers will acknowledge a complete report, investigate it, and
coordinate disclosure and a release when appropriate.

## Scope

Security issues in this provider include credential disclosure, unsafe state
handling, unexpected access outside the configured RevenueCat project, release
artifact tampering, and vulnerabilities in provider code. RevenueCat service or
API vulnerabilities should be reported directly to RevenueCat.
