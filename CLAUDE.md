# CLAUDE.md — user-settings

## What is this project?

A Go service that stores per-user settings and answers privacy questions
about users (today: is this user a minor under their country's age rules).
Twirp RPC over HTTP, DynamoDB for storage, an upstream HTTP user directory
for date of birth and country, OpenTelemetry for tracing.

Read `docs/architecture.md` before changing anything below the API layer and
`docs/development.md` for commands and workflow. This file is the rules.

## ⚠️ Project Invariants — DO NOT VIOLATE

1. **Minor protection fails closed.** If a user's age cannot be determined
   (missing, malformed or impossible date of birth, directory error for that
   user), they are evaluated as age 0 and therefore restricted wherever a rule
   exists. Never "default to adult", never skip the rule check, never swallow
   the parse error silently. The one-month shift in `ageFromDateOfBirth` is a
   deliberate protective margin; do not remove it to make a test tidier.

2. **Errors cross layers as values, not strings.** The API maps domain
   failures to Twirp codes with `errors.Is` / `errors.As` against sentinels
   (`prefsservice.ErrInvalidValue`, `definitions.ErrDefinitionNotFound`).
   Never compare `err.Error()` to a literal, never slice an error message.
   A new failure the API should distinguish gets a new sentinel in the
   package that produces it, wrapped with `%w`.

3. **No package-level state.** `internal/app.New` is the composition root
   and the only place concrete types are constructed together. Handlers,
   services and repositories receive dependencies as constructor arguments
   and hold them as fields. No `init()` side effects, no `sync.Once`
   singletons, no global config or logger lookups (request-scoped loggers
   travel in `context` via `logging.FromContext`, which is the one sanctioned
   context value).

4. **The proto contract is the public API.** Field numbers are never reused
   or renumbered; new fields get new numbers; removing a field means
   `reserved`. Generated code under `internal/generated` is committed and
   must match `rpc/*.proto` (CI checks). Regenerate with `make generate` and
   commit the output in the same change as the proto edit.

5. **Interfaces are declared by the consumer, as small as possible.**
   `domain.UserDirectory` has one method because `domain` calls one method.
   Do not create a shared `interfaces` package, do not widen an interface to
   "match" the concrete type, and do not accept a concrete client where an
   interface would do. Mocks are generated from the consumer-side interface.

6. **Setting values are JSON documents.** `PutSetting` rejects anything that
   does not parse. A string value is therefore `"\"dark\""`, not `dark`.
   Do not relax this at the API edge; downstream consumers rely on it.

If a change needs to violate one of these to work, stop and raise it instead.

## Tech Stack

- **Language**: Go 1.24+, standard library first (`net/http`, `log/slog`,
  `errors`, `slices`, `context`)
- **RPC**: Twirp v8 with protobuf (`rpc/*.proto` → `internal/generated`)
- **Storage**: DynamoDB via `aws-sdk-go-v2`; DynamoDB Local in dev/tests
- **Observability**: OpenTelemetry (`otelhttp`, `otelaws`, `twirpotel`,
  `autoexport`), `slog`
- **Config**: `envconfig`
- **Testing**: `testing` + `testify`, `go.uber.org/mock`, testcontainers,
  WireMock
- **Tooling**: Make, golangci-lint v2, `air` for live reload
- **CI**: GitHub Actions (`.github/workflows/ci.yml`)

## Common Commands

```bash
make help               # list targets
make run                # DynamoDB Local + live-reloading API on :8000
make test               # unit tests, -race, no I/O
make test-integration   # repository tests against DynamoDB Local (make docker-up first)
make test-e2e           # black-box tests; testcontainers starts DynamoDB + WireMock
make test-all           # everything
make lint               # golangci-lint
make generate           # protobuf + Twirp + mocks
make build              # bin/user-settings
go run ./cmd/api        # after sourcing config/.env.local
```

## Project Structure

```
cmd/api/main.go              load config → app.New → httpserver.Run; returns errors, exits non-zero
internal/
  app/                       composition root (the only place concrete types meet)
  api/                       api.Server implements the generated Twirp interface;
                             validation + error mapping only, no business logic
  httpserver/                router, otelhttp, logging middleware, /-/health, graceful Run
  prefsservice/              use cases (GetSettings, PutSetting, GetPrivacySettings, ...)
    domain/                  minor-protection decision: DOB parsing, ageAt, rules
  prefsrepo/                 DynamoDB repository; Setting model; type codes
    definitions/             static catalogue of settings (type, privacy flag, defaults)
  rulesservice/              ApplyRules(age, country)
  regionalrulesrepo/         in-memory per-country rules
  userdirectory/             HTTP client for the upstream directory
  mappers/                   proto <-> prefsrepo.Setting
  config/ logging/ telemetry/ awsdynamo/
  generated/                 protoc output — never edit by hand
rpc/                         .proto files
tests/integration/           repository vs DynamoDB Local
tests/e2e/                   full service over HTTP via httptest + testcontainers
config/.env.*                environment files for run/test/e2e
docs/                        architecture.md, development.md
```

## Architecture Notes

### Layering and dependency direction

`api → prefsservice → (domain, prefsrepo)`. Nothing below `api` imports
Twirp or the generated package. `domain` depends on `userdirectory.User`
(a plain struct) and `rulesservice.RuleApplication`, nothing else concrete.

### Request validation lives in `api`, business rules live in the service

`api` checks shape: session present, URN present, required fields,
known enum. `prefsservice` checks meaning: value is JSON, defaults for
unset privacy settings. Do not duplicate validation across the two (the
original had JSON validation in both the mapper and the service; one was
removed on purpose).

### DynamoDB access pattern

One table, `pk` = user URN, `sk` = setting name. `GetSettings` queries the
partition and filters names in memory: one round trip, cheap, correct for
the handful of settings a user has. If you are tempted to add
`BatchGetItem`, first show the per-user setting count justifies it.
`PutSetting` is an unconditional put (last write wins); add a version
attribute and condition expression if concurrent writers matter.

### Privacy batch fan-out

`GetPrivacySettings` fans out per user with `errgroup` bounded at
`maxConcurrentUserLookups`, writes results by index to preserve order, and
skips (with a warning) users whose lookup fails. Keep all three properties
if you touch it.

### Time

Anything that reasons about "now" takes a `domain.Clock` or an explicit
`time.Time`. Tests pass fixed instants. Do not call `time.Now()` inside
age logic.

## Environment

See `docs/development.md` for the full table. Required: `AWS_REGION`,
`USER_SETTINGS_TABLE_NAME`, `USER_DIRECTORY_BASE_URL`. Set
`DYNAMODB_ENDPOINT` to use DynamoDB Local; set `OTEL_SDK_DISABLED=true`
locally. `config/.env.local` has a working set.

## Conventions

- `gofmt` and `goimports` formatting, enforced by lint
- Package names are short, lower-case, no underscores; file names
  `snake_case.go`; tests next to the code they test
- Exported identifiers have doc comments that start with the identifier name
- Errors are wrapped with context: `fmt.Errorf("loading config: %w", err)`;
  messages are lower-case, no trailing punctuation, no "failed to" prefix
  stacking
- Logging via `slog` with structured attributes (`slog.String("userUrn", u)`),
  never `fmt.Println` or string interpolation into the message
- Mocks are generated (`//go:generate mockgen ...` directives), never
  hand-written

## Code Standards

### Clean Code

- **DRY**: Extract shared logic. If you see duplication, refactor it.
- **SRP**: Every function, type and package does one thing. If a function
  needs an "and" to describe it, split it.
- **Small, composable functions**: Prefer many small functions over few
  large ones. Each should be independently understandable and testable.
- **Never over-engineer**: Write the minimum code that solves the problem
  correctly. No speculative abstractions, no "just in case" interfaces, no
  generics without two real call sites.
- **Naming**: Intention-revealing names. Code should read like prose; a
  comment explains *why*, never *what*.
- **No dead code**: Remove unused identifiers, commented-out code and stale
  comments. Do not leave `TODO`s without an issue link.

## Testing

No PR is mergeable without tests that cover the behaviour introduced or
changed in that PR. This is not negotiable. If it is too hard to test, that
is a signal the code needs restructuring, not that the test can be skipped.

### The Testing Pyramid

```
        /\
       /  \
      / E2E\         tests/e2e      — few; real HTTP, real containers
     /------\
    /  Integ- \      tests/integration — repository against DynamoDB Local
   / ration    \
  /-------------\
 /   Unit Tests  \   internal/**/*_test.go — the bulk; no I/O
/-----------------\
```

**Unit tests** are the base and the majority. No network, no filesystem, no
Docker, no `time.Now()`. They test one function or type in isolation. If a
unit test needs I/O, it is an integration test: move it.

**Integration tests** check a real boundary: today, the repository against
DynamoDB Local. Fewer than unit tests, allowed to be slower, isolated from
each other (each test seeds and clears its own data).

**End-to-end tests** build the app through `app.New`, serve it on an
ephemeral port, and drive it over HTTP. Only for the paths that would be
catastrophic to break silently. They are the last line of defence, not the
first.

### Unit Tests

- Every new function or method gets unit tests. Table-driven where there are
  several inputs; `t.Run` sub-tests with descriptive names:
  `t.Run("unknown date of birth is evaluated as age 0", ...)`, not
  `t.Run("case 3", ...)`.
- Test behaviour, not implementation. A test that breaks when you rename a
  private variable is testing the wrong thing.
- Mock at integration boundaries only (directory client, repository, clock)
  using the consumer-side interfaces. Do not mock your own domain code.
- A test that cannot fail is not a test. After writing one, break the
  implementation and watch it fail (this is how the `NUMBER` bug test was
  written: red first, then the fix).
- Use `require` for preconditions and error checks, `assert` for the
  remaining checks of a case. `testifylint` enforces this.
- Run with `-race`. Concurrency code (`GetPrivacySettings`) must have a test
  that exercises more goroutines than the limit.

### Integration Tests

- New repository methods, new tables or indexes, and new external clients
  get integration tests against the real thing (DynamoDB Local, WireMock).
- Test the contract at the boundary: inputs, outputs, side effects. Not the
  call graph.
- Isolate tests from each other. Shared mutable state between tests is a
  land mine.

### End-to-End Tests

- Any PR that adds an RPC or changes its wire-visible behaviour (status
  codes, error codes, response shape) includes an e2e test.
- Drive the service only through HTTP. Stub upstreams with WireMock
  fixtures in `tests/e2e/wiremock`; never reach into internals.
- No sleeps. Wait on container readiness and HTTP responses.
- Deterministic. A test that passes nine times in ten is broken.

### Coverage

Coverage is a floor, not a goal. Use `make test-coverage` to find untested
paths. New code must not lower coverage.

### What is not an acceptable excuse

- **"It's just a small change."** Small changes break things. Small tests are
  also small.
- **"It's hard to test."** Make it testable. Difficulty is a design signal.
- **"I'll add tests in a follow-up."** You won't. Same PR or no merge.
- **"The existing tests cover it."** Show which ones, in the PR description.
- **"It's generated code."** Then the generator config is the change; test
  what consumes it.

## Security

- Never commit secrets, account IDs, internal hostnames or credentials. The
  original import carried an infrastructure inventory by accident; that is
  the failure mode to watch for. `config/.env.*` hold only local values.
- Validate all input at the API boundary; the generated getters are nil-safe,
  use them. Never dereference a request field without checking.
- Setting values are stored as opaque JSON strings; never evaluate or
  template them.
- The service trusts `user_session`. It is designed to sit behind an
  authenticating gateway. Do not add features that assume otherwise without
  adding authentication.
- Keep dependencies current (Dependabot is configured); run
  `govulncheck ./...` before a release.
- Containers run as non-root on a distroless base; keep it that way.

## Git Workflow

- **Always work from a feature branch.** Never commit directly to `main`.
  Branch names like `feat/batch-get-settings`, `fix/dob-parsing`.
- **Commit often.** Small commits, each one logical unit of work. Each commit
  should build and pass unit tests on its own.
- **Conventional commit messages** with prefixes:
  - `feat:` new capability
  - `fix:` bug fix
  - `refactor:` restructuring with no behaviour change
  - `perf:` performance
  - `test:` tests only
  - `docs:` documentation
  - `chore:` build, CI, dependencies, tooling
  - `ci:` workflow changes
- **Messages describe what and why, not how.** The body explains the
  reasoning a reviewer would otherwise have to reconstruct. Example:
  `fix: treat an unknown date of birth as a minor instead of an adult`, not
  `update process_setting.go`.
- **Submit PRs to `main` with `gh pr create`.** Same prefixes in the title.
- **The user reviews every PR before merge.** Do not merge autonomously.
- **NEVER add `Co-Authored-By`, "Generated with Claude Code", session links
  or any other AI attribution to commits, PRs or code comments.**

### PR Description Template

```markdown
## Scope

<!-- WHAT and WHY. The big picture. -->

closes #<issue>

## Implementation

<!-- HOW. Program flow, refactors, trade-offs, where to look closely. -->

## How to Test

<!-- 1) Automated: which unit / integration / e2e tests cover this.
     2) Manual: exact commands (curl against make run) to see it. -->

## Invariants & Risk

<!-- Does this touch minor protection, error mapping, the proto contract,
     the DynamoDB key schema or generated code? Confirm the Project
     Invariants hold. Call out any config or schema change. -->

## Emoji Guide

| Type         | Emoji          | Meaning                                       |
| ------------ | -------------- | --------------------------------------------- |
| Blocking     | 🔴 ❌ 🚨       | Must be addressed before merge                |
| Non-blocking | 🟡 💡 🤔 💭    | Minor suggestion, nit, or clarifying question |
| Praise       | 🟢 💚 😍 👍 🙌 | Positive feedback, a crucial part of review   |
```

## Go

### Tooling & CI Enforcement

- `gofmt` / `goimports` clean, `go vet` clean, `golangci-lint run` clean.
  All three run in CI; none are optional.
- `go test -race` for unit tests. Data races are bugs, not flakes.
- Generated code is committed and CI verifies it is fresh.
- Keep the `go` directive in `go.mod` current and dependencies tidy
  (`go mod tidy` leaves no diff).

### Error Handling

- Return errors; do not log and return. The caller decides what to log.
  Log once, at the boundary that handles the error (the API layer, or the
  place that deliberately swallows it, with a reason).
- Wrap with context and `%w`: `fmt.Errorf("creating table %s: %w", name, err)`.
  The chain should read like a stack: outermost context first.
- Sentinel errors (`var ErrX = errors.New(...)`) for conditions callers
  branch on; custom error types only when the caller needs structured data
  from the error. Compare with `errors.Is` / `errors.As`, never `==` on
  wrapped errors and never on message text.
- No `panic` in library code. `main` returns an error from `run()` and exits
  non-zero. A failed startup is an error with context, not `log.Fatal` buried
  three packages deep.
- Check every error. `_ =` is acceptable only with a comment saying why the
  error cannot matter (closing a response body after reading it fully).

### Context

- `context.Context` is the first parameter of anything that does I/O or can
  block, named `ctx`. Never store it in a struct.
- Propagate the caller's context; derive with `WithTimeout` /
  `WithCancel` and always `defer cancel()`.
- The only values in context are request-scoped cross-cutting concerns
  (logger, trace). Dependencies travel as arguments or fields.

### Interfaces & Types

- Accept interfaces, return structs. Declare interfaces where they are
  consumed, with only the methods used (see Invariant 5).
- Assert interface satisfaction at compile time when a type exists to
  implement one: `var _ proto.UserSettingsService = (*Server)(nil)`.
- Prefer value receivers for small immutable types, pointer receivers when
  the type has mutable state or is large. Be consistent within a type.
- Typed constants with `iota` get an explicit type
  (`RuleAgeRestriction RuleType = iota`) and a `String()` method if they
  are ever logged or compared to text.
- Zero values should be useful. A struct that must be constructed through a
  function should be unexported or documented as such.

### Concurrency

- Goroutines have a clear owner and a clear end. Use `errgroup` (with
  `SetLimit` for fan-out) or `sync.WaitGroup`; never fire-and-forget.
- Bound all fan-out. Unbounded goroutines against a database are a
  self-inflicted outage.
- Write results by index into a pre-sized slice to preserve order without a
  mutex; use a mutex when you must append.
- Channels are for communicating ownership; a mutex is fine for guarding a
  map. Pick the simpler one.
- Every concurrent path has a `-race` test with more work items than the
  concurrency limit.

### HTTP & Twirp

- Set `ReadHeaderTimeout` (and others as needed) on every `http.Server`.
- Outbound clients carry a timeout or take a context with a deadline; always
  close response bodies; check status codes before decoding.
- Validation and error mapping are the handler's whole job. Business logic
  does not belong in `api`.
- Use the generated nil-safe getters (`req.GetUserSession().GetUserUrn()`),
  never raw field access on request messages.

### Control Flow & Style

- Early return over nested `if`. Happy path at the lowest indentation.
- `switch` over `if/else` chains with three or more arms; include a
  `default` that does the safe thing.
- Prefer the standard library: `slices`, `maps`, `strings`, `errors`,
  `log/slog`. Reach for a dependency when it carries real weight
  (AWS SDK, testcontainers), not for a helper you could write in ten lines.
- Pre-size slices and maps when the length is known
  (`make([]T, 0, len(xs))`).
- No `init()` functions. No package-level `var` with side effects.
- Comments explain intent and non-obvious trade-offs (`// fail closed:
  unknown age is 0`). Delete comments that restate the code.

### Testing Idioms

- Table-driven tests with named cases; `t.Run` for each; `t.Parallel()`
  where cases share no state.
- `t.Helper()` in every test helper so failures point at the caller.
- `t.Cleanup` over `defer` for teardown in helpers.
- Fixed clocks and fixed inputs; no `time.Now()` in assertions, no random
  data without a logged seed.
- Mocks via `mockgen` from the consumer interface. Set expectations only on
  calls the behaviour under test actually makes; `gomock.Any()` for
  arguments that are not the point of the test.

### Iteration Workflow

- Get it working first with the straightforward code, then tighten: narrow
  interfaces, remove duplication, add the error context, run the linter.
- When the compiler or the race detector fights you, the design is telling
  you something. Fix the design, do not suppress the tool.

### Reviewer Empathy

- Formatting, vet and lint clean so the reviewer spends attention on what
  the code does.
- Small PRs with a commit history that tells a story. If a reviewer has to
  reconstruct why, the message or the comment is missing.
- Every behavioural change is visible in a test diff.
