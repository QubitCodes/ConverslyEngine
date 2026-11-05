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

Implementation commit, deployed image digest and runtime acceptance
must be entered after release. The Conversly automatic-setup
adapter includes a pre-mutation read-only permission gate; public provisioning
stays disabled until this patch and live isolation acceptance pass.

## Upgrade and recovery

On each upstream upgrade compare all five handlers and the global middleware.
Retain the patch unless upstream supplies equivalent checks and the regression
tests pass on the candidate. Recheck exact tenant-key identity and forbidden role
access after deployment. A code rollback restores the vulnerable behavior, so
disable new customer onboarding/provisioning before any rollback. Do not erase
provisioning journals or change the owner's MSG91/phone setup as recovery.
