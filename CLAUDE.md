# CLAUDE.md — user-settings

Go service: per-user settings and "is this user a minor" answers. Twirp over HTTP,
DynamoDB, an upstream HTTP user directory, OpenTelemetry. Read `docs/architecture.md`
before touching anything below `internal/api`; `docs/development.md` for workflow;
`terraform/README.md` for infrastructure. This file is only what you cannot infer
from the code.

## Boundaries — do not cross without raising it first

1. **Minor protection fails closed.** Unknown or unparseable date of birth evaluates as
   age 0 (restricted wherever a rule exists). Never default to adult, never skip the
   rule check. The one-month shift in `ageFromDateOfBirth` is deliberate; keep it.
2. **Errors cross layers as values.** Sentinels (`prefsservice.ErrInvalidValue`,
   `definitions.ErrDefinitionNotFound`) wrapped with `%w`; the API maps them with
   `errors.Is`. Never compare or slice `err.Error()`.
3. **No package-level state.** `internal/app.New` is the only composition root.
   Dependencies are constructor arguments held as fields. No `init()`, no singletons,
   no global config. The request logger in `context` is the one sanctioned context value.
4. **The proto contract is the public API.** Never reuse or renumber a field; removed
   fields become `reserved`. `internal/generated` is committed and must match
   `rpc/*.proto` (CI checks). Regenerate and commit in the same change.
5. **Interfaces are declared by the consumer, as small as possible.** No shared
   `interfaces` package; mocks are generated from the consumer-side interface.
6. **Setting values are JSON documents.** `PutSetting` rejects anything else. A string
   value is `"\"dark\""`. Do not relax this.
7. **Nothing account-specific is committed.** No AWS account IDs, ARNs, hostnames or
   secrets, in code, Terraform, workflows or fixtures. The original import leaked an
   infrastructure inventory; that is the failure mode to watch for.
8. **Terraform consumes tf-tools; it does not fork it.** Missing capability goes into
   `github.com/octamagic/tf-tools` as a module or an additive flag, then gets
   pinned here. Every `?ref=` moves together.

## Commands you cannot guess

```bash
make help                 # the full list
make test                 # unit only; packages under tests/ are excluded on purpose
make test-integration     # needs `make docker-up` (DynamoDB Local on :8900, not 9000)
make test-e2e             # testcontainers; needs Docker, nothing else
make generate             # protoc + Twirp + mockgen; commit the output
make tf-test              # validate every Terraform root + `terraform test`, against a
                          # LOCAL tf-tools checkout (TF_TOOLS_DIR, default ~/code/tf-tools)
scripts/tf-local.sh       # what tf-test runs; prints the temp dir it worked in
/add-rpc <Name>           # skill: proto → generate → handler → tests, in order
/tf-bump <tag>            # skill: move every tf-tools pin, refresh lock files, revalidate
```

- Test and e2e env files are sourced from `config/.env.*`; the Makefile does this.
  Running `go test` by hand on `tests/e2e` without `set -a; . config/.env.e2e` fails
  on config load.
- `make generate-proto` runs `protoc` **twice** (once for `--go_out` over both protos,
  once for `--twirp_out` over the service proto). protoc-gen-twirp refuses files with
  different `go_package` values in one invocation. Do not "simplify" it into one call.
- tflint is not in Homebrew on this machine. Use the binary from the tflint GitHub
  release; `make tf-lint` expects it on `PATH`.
- Golden local check before a commit: `go vet ./... && golangci-lint run && make test`.

## How this repo differs from the default

- **Tests are the spec and come first.** A bug fix starts with the failing test
  (the NUMBER-mapping fix is the model). A `terraform test` assertion that cannot
  fail is deleted; break the module on purpose once to prove it.
- **Pyramid, enforced by location.** `internal/**` unit tests: no I/O, fixed clocks,
  mocks only at boundaries (directory client, repository, `domain.Clock`).
  `tests/integration`: repository vs DynamoDB Local. `tests/e2e`: over HTTP only,
  upstreams stubbed with WireMock fixtures in `tests/e2e/wiremock`. Nothing merges
  without tests for the behaviour it changes; "follow-up PR" is not an answer.
- **`require` for errors and preconditions, `assert` for the rest.** testifylint
  enforces it; `golangci-lint run --fix` handles most of the churn.
- **Validation lives in `api`, rules live in the service.** Shape checks (nil
  session, empty name) at the edge with the nil-safe generated getters; meaning checks
  (valid JSON, defaults) in `prefsservice`. The original duplicated JSON validation in
  two layers and one was removed on purpose.
- **Fan-out is bounded and ordered.** `GetPrivacySettings` uses an errgroup with a
  limit, writes results by index, and skips (with a warning) users whose lookup fails.
  Any concurrent path gets a `-race` test with more items than the limit.
- **Deploys do not fight Terraform.** CI rolls images by registering ECS task
  definition revisions; the service module sets `track_latest_task_definition = true`
  so the next apply follows the live revision. `image_tag` only matters on first apply.
- **Terraform tests assert plan-time-known values only.** Under `mock_provider`
  computed attributes (ARNs, URLs) are unknown at plan and random at apply, and the
  AWS provider rejects random strings where it validates ARN shapes. Keep assertable
  data in `locals` (`local.environment`, `local.container`, `local.ci_trust_policy`).
- **Workflows are composed, not copied.** `.github/workflows/*` call
  `octamagic/actions-toolkit` reusable workflows pinned to `@v0` (there is no
  `v1` tag yet, whatever the toolkit's docs say). Credentialed jobs gate on
  repository variables so a fresh clone stays green.
- **Git.** Feature branches only; conventional commit prefixes (`feat`, `fix`,
  `refactor`, `perf`, `test`, `docs`, `chore`, `ci`); the body says why. Each commit
  builds and passes unit tests on its own. PRs via `gh pr create` with Scope /
  Implementation / How to Test / Invariants & Risk sections; the user merges.

## Gotchas that have bitten more than once

- **Ports 8000 and 9000 are taken on this machine** by other Docker containers. The
  service and e2e tests use 8000 only inside containers or `httptest`; DynamoDB Local
  is published on 8900. Do not move them back.
- **Stale `.terraform.lock.hcl` after a provider bump.** `scripts/tf-local.sh` copies
  the roots including lock files; a constraint change to `~> 6.0` then fails `init`
  with "Failed to query available provider packages". Delete the lock files, rerun,
  copy the regenerated ones back.
- **AWS provider 6 renamed `data.aws_region.current.name` to `.region`.** Roots pin
  `~> 6.0` and the module requires `>= 6.0`; do not reintroduce `.name`.
- **BSD sed.** `\|` alternation and `\s` do not work in macOS `sed`/`awk`; use `perl`
  for anything beyond a literal substitution, or edit the file directly.
- **Twirp URL paths.** The route is `/twirp/usersettings.v1.UserSettingsService/<Rpc>`;
  the generated `UserSettingsServicePathPrefix` already ends with the service name.
- **The toolkit's `.toolkit` checkout** appears in the workspace during CI runs. It is
  gitignored here; a glob over the repo in a build step must skip it.
- **Generated Twirp bytes depend on the Go version that built protoc-gen-twirp.** It
  gzips the descriptor and compress/flate output moves between Go releases. `make tools`
  builds the generators with the go.mod Go version for that reason; a diff that is only
  the gzipped `FileDescriptorProto` means the tools were built with a different Go.
- **A DynamoDB attribute listed twice** (as a key and again in `attributes`) blew up
  the `for` expression in the tf-tools module until it grouped with `...`. Dedup by
  name is required and tested there.

## Go conventions in one screen

- Return errors, wrap with context (`fmt.Errorf("creating table %s: %w", name, err)`),
  log once at the boundary that handles them. No `panic` outside `main`'s `run()`.
- `ctx` first on anything that blocks; never stored in a struct.
- Accept interfaces, return structs; assert satisfaction at compile time
  (`var _ proto.UserSettingsService = (*Server)(nil)`).
- Typed `iota` constants with a `String()`; early returns; `switch` with a safe
  `default`; pre-size slices; standard library before a dependency.
- `slog` with attributes, never string interpolation into the message.
- Exported identifiers carry a doc comment starting with their name; comments
  explain why, never what. No dead code, no TODO without an issue.
- gofmt/goimports, `go vet`, golangci-lint (config in `.golangci.yml`) and `-race`
  are all CI gates; none are optional.

## Where things are

```
cmd/api/            run() returns errors; signal.NotifyContext; graceful drain
internal/app        composition root        internal/api        Twirp handlers + error mapping
internal/httpserver router, middleware      internal/prefsservice  use cases (+ domain/ for age rules)
internal/prefsrepo  DynamoDB + definitions  internal/userdirectory upstream HTTP client
internal/generated  protoc output, committed, never edited
terraform/          tf-tools composition; docs in terraform/README.md
.github/workflows   ci, terraform, deploy (actions-toolkit); claude-code-review, claude (gated on CLAUDE_REVIEWS)
```
