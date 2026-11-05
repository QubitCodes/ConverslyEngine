# WM-001: custom-role authorization

Approved: 5 November 2026. Base:
`047127ce47d3ba631aa66da9bd62b2cee8bc2c68`. Engine fork:
https://github.com/QubitCodes/ConverslyEngine.

## Reason and root cause

The global API middleware authenticates requests and explicitly delegates granular
permissions to handlers. Five role handlers instead checked only organisation
identity. A restricted holder of a custom role could edit that role's permission
list. Existing resource UUID scoping does not prevent this within-tenant privilege
escalation. No production exploitation was attempted or observed.

## Exact intended change

`internal/handlers/roles.go`: ListRoles and GetRole require `users:read`;
CreateRole, UpdateRole and DeleteRole require `users:write`, using the existing
organisation-aware `requireAuth` helper. Existing Super Admin handling and
immutable-system-role restrictions stay intact. No schema/data migration,
credential change, port exposure, frontend feature or dependency upgrade.

The [source patch](WM-001-role-authorization.patch) passes `git apply --check`
against the base checkout. Regression tests must prove restricted reads/writes
are 403, self-role escalation does not change stored permissions, an authorised
administrator still succeeds, and cross-organisation resources remain isolated.

## Release record

The five-handler implementation and regression tests are present. All five
restricted-access cases fail against the original base (as expected); they pass
with this patch. Role/permission and organisation tests pass against a dedicated
loopback PostgreSQL database and isolated Redis; `go vet ./internal/handlers`
passes. Test fixtures that exercise authorised role management now have explicit
user-management permissions, without making them Super Admin. No test used
production services. Go 1.26.0 matches the existing Dockerfile toolchain.

Source files changed: `internal/handlers/roles.go`, `roles_test.go` and the new
`roles_authorization_test.go`. No dependency or schema changes.

Implementation commit: `ed7a9533b9ed902200e07cc7a21061d97e622b74`, pushed to
the owned engine fork's main branch. Production deployment completed on 5 November 2026.
Engine image: `sha256:4cd71929db7be5dbfaa187b78e7d75549e80f271528cd7f82a537cf0c4aeb571`. The Conversly automatic-setup
adapter includes a pre-mutation read-only permission gate; public provisioning
stays disabled until this patch and live isolation acceptance pass.

## Upgrade and recovery

On each upstream upgrade compare all five handlers and the global middleware.
Retain the patch unless upstream supplies equivalent checks and the regression
tests pass on the candidate. Recheck exact tenant-key identity and forbidden role
access after deployment. A code rollback restores the vulnerable behavior, so
disable new customer onboarding/provisioning before any rollback. Do not erase
provisioning journals or change the owner's MSG91/phone setup as recovery.

## Production verification - 5 November 2026

Engine, web and worker are healthy on their expected commits. Web and worker use
`2d5434a48e5c1f181aa76a45c84beae013e786a5`. All 21 application migration
checksums match the release SQL, including additive migration 0020. Strict TLS
to the Google origin with the public hostname returns 200 for `/login`.

The existing Qubit Codes service identity is not Super Admin. Its saved role
explicitly includes `users:read`; roles/users reads therefore correctly return
200, while API-key and organisation administration return 403. Its role has no
`users:write`. This existing identity is not the narrower new provisioner and was
not broadened or replaced during these checks. The separately approved control identity and two-tenant operator acceptance now
pass (see below). Public Meta onboarding remains disabled.

Root-only proof: `conversly-google-release-20261105/wm001-release-acceptance.json`.
Pre-release application and engine dumps were validated with pg_restore listing.
No existing phone number, conversation or template was moved.

## Separate control and tenant acceptance

Approved restricted control organisation: `5c3a78f3-457f-44b2-a40a-f412cf713852`.
The non-Super-Admin provisioner has only `organizations:write` in that empty home
organisation; roles/users/API-key/organisation listing all return 403. The key is
saved encrypted only in the private worker. The existing tenant key is unchanged.

The deployed provisioning adapter created two explicitly approved empty operator
acceptance tenants, using separate random acceptance identifiers rather than paid
customer records. No production order or subscription was fabricated.

- A: `11b72fa6-9d3b-4af4-91c6-ee519e862f5e`.
- B: `584497c1-816e-4e83-a2c4-2db5ab76fe78`.

Each has a different non-Super-Admin technical user and encrypted one-time key.
Final permissions exactly match `tenantEnginePermissions`; temporary key-creation
permission was removed. Each tenant's roles/users/API-key/organisation listing is
403. Foreign organisation headers cannot access the other tenant. The provisioner
cannot access the pre-existing Qubit Codes tenant. No business messages, numbers,
contacts or templates were added to these test organisations. Keep their IDs in
this register; do not silently adopt them for customer workspaces.

Root-only host receipts: `tenant-isolation-acceptance.json`,
`tenant-isolation-A.json`, `tenant-isolation-B.json` under the release backup
folder. Their checkpoint keys are encrypted with the acceptance workspace context.
The operator procedures are retained alongside this record. Never replay an
operation with an existing journal; reconcile it first.
