# Whatomate patch register

Record every local engine change here before release. This is the upgrade ledger,
not a competing product checklist. The user requested this folder and approved
the narrow role-authorization patch on 5 November 2026.

| ID | Purpose | Upstream base | State | Evidence |
| --- | --- | --- | --- | --- |
| WM-001 | Require user-management permissions for role reads/writes; prevent custom-role self-escalation | `047127ce47d3ba631aa66da9bd62b2cee8bc2c68` | Implemented; regression checks and Go vet passed; deployment pending | [Patch](WM-001-role-authorization.patch), [record](WM-001-role-authorization.md) |

For each upgrade, compare the proposed upstream code with every retained patch.
Keep a patch until an equivalent upstream fix and the same regression tests pass.
Do not assume a clean merge means the security behavior survived. Record the new
upstream SHA, patch/fork commit, tested image digest, migration implications and
verified release result. Never silently drop a patch or enable automatic upstream
promotion; the established staging/release gate remains applicable.

Keep credentials, customer conversations and private connection strings out of
this folder. Include only source diffs, redacted checks and operational identities.
