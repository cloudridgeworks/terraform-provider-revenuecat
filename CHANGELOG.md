## 0.1.0 (Unreleased)

FEATURES:

* Add entitlement, offering, package, webhook, entitlement-product, and package-product resources.
* Add local validation for RevenueCat API v2 identifiers, names, webhook configuration, package positions, and eligibility criteria.

BUG FIXES:

* Omit package position when it is not configured so RevenueCat can choose its API default.
* Handle webhook responses that omit the computed signing secret.
