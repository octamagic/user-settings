# Architecture

This document explains how the service is put together and the reasoning
behind the main decisions. Read it before changing anything below the API
layer.

## Shape

```
          HTTP (Twirp, JSON or protobuf)
                    │
   httpserver ──────┤  otelhttp → logging middleware → Twirp router, /-/health
                    │
   api.Server ──────┤  validate request, call service, map errors to Twirp codes
                    │
   prefsservice ────┤  use cases: get/put settings, privacy settings with defaults
        │           │
        ├─ domain ──┤  minor-protection decision: parse DOB, compute age, apply rules
        │     ├── userdirectory (HTTP)      upstream system of record for DOB/country
        │     └── rulesservice → regionalrulesrepo   per-country age thresholds
        │
   prefsrepo ───────┤  DynamoDB table (pk = user URN, sk = setting name)
        └─ definitions   static catalogue of known settings and their defaults
```

Dependencies point inward: `api` knows `prefsservice`, `prefsservice` knows
`prefsrepo` and `domain`, and nothing below knows about Twirp or HTTP.

### Composition root

`internal/app.New` is the only place concrete types meet. It builds the
DynamoDB client, the user directory client, repositories, services and the
HTTP handler, and returns an `*App` or an error. `cmd/api` calls it and runs
the server; the e2e suite calls it and serves the handler on an `httptest`
port. There is no package-level state anywhere in the service.

### Interfaces are declared where they are consumed

Go interfaces are satisfied implicitly, so a consumer should declare the
smallest interface it needs rather than import a fat one from the provider.
For example `domain` declares `UserDirectory` with one method and
`RulesService` with one method; `prefsservice` declares its own
`SettingsRepository` with the two methods it calls. The concrete types in
`userdirectory`, `rulesservice` and `prefsrepo` satisfy them without knowing
they exist. Mocks are generated from the consumer-side interfaces.

`domain.Clock` exists so age calculations can be tested at a fixed instant.

## Request flow: `PutSetting`

1. `api.Server.PutSetting` extracts the caller's URN with nil-safe generated
   getters. A missing session or URN is `InvalidArgument`.
2. The mapper converts the proto setting to `prefsrepo.Setting`. An unknown
   type enum is `InvalidArgument`.
3. `prefsservice.PutSetting` checks that the value is a JSON document and
   returns `ErrInvalidValue` (wrapped with `%w`) if not.
4. `api` maps `errors.Is(err, prefsservice.ErrInvalidValue)` to
   `InvalidArgument` and anything else to `Internal`.

Error mapping is by sentinel value, never by message text. If you add a new
failure mode the API should distinguish, add a sentinel in the package that
produces it and an `errors.Is` branch in `api`.

## Request flow: `GetPrivacySettings`

1. Ask `domain` for each user's privacy attributes (currently: is minor).
2. Fan out one DynamoDB query per user, bounded at ten in flight, keeping
   request order. Each user's partition is independent, so there is nothing
   to gain from a single large query.
3. For settings the user never set, synthesise values from the definition's
   defaults. Definitions can carry a separate default for minors.
4. A user whose lookup fails is logged and omitted; the batch still succeeds.

## Minor protection

`domain.minorSettingsProcessor` decides whether a user is a minor:

- Date of birth arrives as `YYYY/MM/DD` from the directory. Parsing is strict:
  missing, non-numeric, or impossible dates (`2000/02/31`) are errors.
- **Unknown age is treated as 0.** This is the fail-closed choice for a
  compliance path: a user we cannot age is restricted wherever a rule exists,
  and unaffected where no rule exists. The previous implementation fell
  through to year 0001, an age of about two thousand, and no restriction.
- Age uses calendar arithmetic (`ageAt`), not `hours / 24 / 365`.
- The date of birth is shifted one month later before computing age because
  the upstream day-of-month is not always reliable. Around a birthday this
  makes the user count as younger, which errs toward protection. Tests pin
  this behaviour.
- `rulesservice.ApplyRules(age, country)` returns `ApplyRestriction` when a
  rule for the country has a threshold above the age.

## Storage

One DynamoDB table:

| attribute | role | notes |
| --- | --- | --- |
| `pk` | hash key | user URN |
| `sk` | range key | setting name |
| `type` | | `s`, `b` or `n`; mapped to the proto enum by `prefsrepo.ParseSettingType` |
| `value` | | JSON document as a string |
| `updatedAt` | | unix seconds |

`GetSettings` runs a key-condition query on `pk` and filters by `sk` in
memory. With a handful of settings per user this is one round trip and one
read unit; a `BatchGetItem` would cost the same or more. If the per-user
setting count grows into the hundreds, switch to `BatchGetItem` for explicit
names and keep the query for "all settings".

`PutSetting` is an unconditional `PutItem`, so last write wins. There is no
version attribute; add one with a condition expression if concurrent writers
become a concern.

## Observability

- `log/slog`, JSON in production and text locally (`LOG_PRETTY`). The HTTP
  middleware attaches the logger to the request context; lower layers fetch it
  with `logging.FromContext` so request-scoped fields propagate.
- OpenTelemetry traces on inbound HTTP (`otelhttp`), the Twirp layer
  (`twirpotel`), outbound HTTP to the directory and every AWS SDK call.
  The exporter is chosen by the standard `OTEL_*` variables via `autoexport`;
  `OTEL_SDK_DISABLED=true` turns it off.
- `/-/health` returns `OK` for load balancer checks.

## Shutdown

`main` uses `signal.NotifyContext`. On SIGINT/SIGTERM the HTTP server drains
in-flight requests for up to ten seconds, then telemetry is flushed. Listen
failures are returned as errors, not panics, so a bad port exits non-zero
with a log line.

## What changed from the original and why

The commit history has the detail; in summary:

| area | before | after |
| --- | --- | --- |
| dependencies | private platform library for config, logging, OTel, HTTP client, session proto | small local packages; builds from a clean clone |
| wiring | package globals, `sync.Once` singleton, `InitService` | `app.New` returns the graph; handlers hold fields |
| errors | `err.Error() == "..."` and a `[0:13]` slice that could panic | sentinel errors and `errors.Is` |
| numeric settings | reported as `STRING` | `NUMBER` (mapper bug) |
| unknown date of birth | treated as an adult | treated as age 0 |
| age math | hours / 365 | calendar arithmetic |
| privacy batch | serial N+1 | bounded fan-out, order preserved |
| dead code | several unreferenced exported functions | removed |

## Deployment

```
GitHub Actions ──OIDC──▶ per-environment deploy role ──▶ ECR push, ECS roll
                                      │
Terraform (tf-tools modules) ─────────┴──▶ ECR · DynamoDB · ECS cluster + Fargate service
                                           ALB (internal) · autoscaling · logs · IAM
```

- One composition module, three environment roots that differ only in sizing.
- The task role can touch exactly one table. The execution role has the AWS managed
  ECS execution policy. The deploy role can push to one repository, register task
  definitions, update one service, and pass the two task roles to ECS, nothing else.
- Image rollouts register a new task-definition revision; Terraform follows the newest
  revision (`track_latest_task_definition`) so the two do not fight.
- Tasks run on private subnets with a read-only root filesystem, reachable only from
  the load balancer's security group on the service port.
