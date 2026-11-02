# Credential-management Coding Plan query contract

2026-11-02. SPEC92 minimal Qwen component slice; **not full functionality, deployed integration, or ownership verification**.

## Canonical classification and scope

CPA credential management means **Coding Plan**. The AI-provider entry means prepaid PAYG cash (out of scope here). Provider name, inference hostname and successful `/models` discovery cannot override this classification or prove ownership. Inference routing is unchanged by this patch.

This plugin's credential quota command accepts only the versioned Coding Plan helper protocol. Explicit legacy `bsk`, `cookie`, `bl`, Token Plan cache/check modes, and missing/default product configuration fail as `needs_configuration` **before subprocess execution**. Legacy standalone tooling remains for compatibility, not as an alternative credential query or optional enrichment. Do not configure it for this lane.

## Explicit read-only query

Example arguments for the repository helper (documentation only; no real token/config is created by this change):

```text
bailian-quota --protocol qwen-coding-plan-v1 --source coding-plan --token-file <protected-raw-console-token-file> --console-region cn-beijing --console-site domestic --timeout 90s --json
```

The other allowed pair is `ap-southeast-1` / `international`. Unknown combinations fail; there is no hostname/default-region inference. The file must contain a raw, pre-enrolled console bearer token, not a CLI profile/login JSON. Protect it outside the repository, readable only by the intended query identity. Missing/expired tokens require separate enrollment outside the query path. This task neither performs nor authorizes enrollment.

The helper does not execute `bl`, `bsk`, login, token refresh, key creation, a browser, or a default-profile lookup. It does not write token/config/cache files. `--cache-only`, `--check`, browser/cookie/BL flags, or an alternate source are incompatible even when supplied with false/empty values. No raw token is supplied in argv, diagnostics or output; the selected inference key is never passed to this helper.

One fixed HTTPS POST query is allowed, with system certificate verification and no redirect:

| Region / site | Origin | Action | Commodity |
| --- | --- | --- | --- |
| cn-beijing / domestic | `https://bailian-cs.console.aliyun.com` | `BroadScopeAspnGateway` | `sfm_codingplan_public_cn` |
| ap-southeast-1 / international | `https://bailian-singapore-cs.alibabacloud.com` | `IntlBroadScopeAspnGateway` | `sfm_codingplan_public_intl` |

Path `/cli/api.json`; query `product=sfm_bailian`, `action=<above>`, `api=zeldaEasy.broadscope-bailian.codingPlan.queryCodingPlanInstanceInfoV2`. Form body contains only `params=<JSON>` and `region=<explicit region>`. Header `Authorization: Bearer <explicit console token>`; no inference bearer/cookie or account-switch selector.

`params` is `{Api:<operation>,V:"1.0",Data:{queryCodingPlanInstanceInfoRequest:{commodityCode:<above>,onlyLatestOne:true},cornerstoneParam:{protocol:"V2",console:"ONE_CONSOLE",productCode:"p_efm",switchUserType:3,consoleSite:"BAILIAN_ALIYUN"}}}`. No arbitrary URL/operation/body extension is accepted.

## Query result versus selected-key authority

Our helper protocol is **not the official CLI JSON**. Example of the first window (all three are required):

```json
{
  "protocol":"qwen-coding-plan-v1",
  "product":"coding-plan",
  "queryStatus":"success",
  "authorityStatus":"needs_configuration",
  "reason":"account_authority_unverified",
  "observedAt":"2026-11-02T12:00:00Z",
  "instanceType":"pro",
  "windows":[
    {"window":"5h","unit":"requests","used":"38","total":"100","remaining":"62","resetTime":"2026-08-06T07:06:40Z"},
    {"window":"week","unit":"requests","used":"500","total":"1000","remaining":"500","resetTime":"2026-08-07T10:53:20Z"},
    {"window":"billing_month","unit":"requests","used":"950","total":"1000","remaining":"50","resetTime":"2026-08-08T14:40:00Z"}
  ]
}
```

- Select the first `status == VALID` instance, as the frozen first-party source does. `instanceType` is a tier, **not a unique plan ID**. No VALID instance fails without guessing expired/absent/mismatched-account status.
- Native counts are **model requests**, not tokens, money or percentages. Require all three windows. Used >= 0, total > 0, used <= total; no clamp, rounding, unlimited/zero inference, percentage back-calculation or partial policy reading.
- Decode upstream JSON numbers without float conversion; accept only unsigned integer lexical form, bounded to 64 decimal digits. Serialize exact counts as decimal strings. Missing/null/strings/fractions/exponents or contradictory remaining counts fail closed. The 64-digit bound is an adapter resource limit, not a claimed official maximum.
- Millisecond epochs must be integers in [946684800000, 253402300799999]. Preserve milliseconds in RFC3339Nano; reject seconds/fractions rather than converting units. Billing-month reset is not computed as 30 days. Reset passage does not locally restore quota.
- `observedAt` is helper receive/parse time, not an official server sample timestamp or freshness/authority proof. There is no new cache, freshness TTL or recovery mechanism in this slice.
- A single complete JSON value, bounded response and output, gateway/business/nested success/error checks, nonzero-exit rejection, existing command deadline and diagnostic redaction remain mandatory. New code does not inherit the legacy BL Windows teardown-abort exception.
- The plugin independently validates the versioned result, counts, unit and three-window identity. Unknown fields, helper `verified` claims or a changed authority state are rejected. Neither selected-key attributes nor hand-entered account/login JSON can bypass the guard.

`Manager.FetchQuota` still returns **empty quota + `needs_configuration: account_authority_unverified`** after a parsed query. It returns no groups, subscription or summary usable for pause/recovery. The plugin management `quota-usage` route retains the selected credential card and adds `queryStatus:success|failed` and `authorityStatus:needs_configuration`; its windows/metrics stay empty and its error explains the missing authority. One query's status may appear on several cards, but its account quota is never copied to those keys. No Core guard is removed or changed.

## Public protocol evidence

Frozen source: first-party `modelstudioai/cli@a7cb810c68f434c2c7330d8c0775614f7e699c8d`, package 2.7.2. These are documentation/source references, **not installed or executed dependencies**:

- `packages/commands/src/commands/usage/coding-plan.ts:5-11,49-67,107-112`: operation, commodity, first VALID instance, `onlyLatestOne`.
- `packages/core/src/console/gateway.ts:17-41,90-112,127-161`: fixed origins/action, form envelope and bearer header.
- `packages/commands/tests/coding-plan-usage.test.ts:68-78,169-221`: three-window fixture, incomplete field semantics. Independent expected remaining counts: 62 / 500 / 50.
- `packages/core/src/client/client.ts:252-290` and `packages/core/src/auth/refresh-token.ts:124-156`: why directly invoking official CLI cannot be asserted strictly read-only.

Public copies/manifests and the canonical research contract are under `D:/WILL/AGENT/CPA/ops/implement-spec92-20261102/development-completion/token-plan-public` and `D:/WILL/AGENT/CPA/ops/implement-spec92-20261102/development-completion/token-plan-contract.md`.

## Evidence and finite outstanding work

Tests exercise exported `bailianquota.RunCLI`, plugin `Manager.FetchQuota` and plugin `Manager.HandleManagement` with real helper subprocesses and a controlled HTTP boundary. Wrong Token Plan + optional Coding Plan mixed requests were observed in the original RED. This is **supplementary component evidence**, not Manager -> SQLite -> actual Core end-to-end evidence. The local Core checkout is used only as a test module replacement for SDK compilation; its guard/sources are not modified.

Evidence and exact counts/terminal receipts: `D:/WILL/AGENT/CPA/ops/implement-spec92-20261102/development-completion/qwen-coding-implementation`.

Still needed before policy-enabled completeness: (1) a trustworthy official selected inference key -> exact account/Coding Plan instance -> query principal binding contract and independent verifier; (2) host authority integration without trusting helper booleans, followed by the agreed Manager/SQLite/real-Core controlled-upstream acceptance seam; (3) separate, explicitly authorized enrollment and live acceptance if desired. This is a missing verified binding implementation/contract, **not merely a request for credentials**. BSS balance, deployment and UI work are not part of this patch.
