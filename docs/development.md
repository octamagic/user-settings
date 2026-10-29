# Development

## Prerequisites

- Go 1.24 or newer (see `go.mod`)
- Docker, for DynamoDB Local and the test containers
- `protoc` (`brew install protobuf`), only if you edit `rpc/*.proto`
- `golangci-lint` v2 (`brew install golangci-lint`)
- `make tools` installs `protoc-gen-go`, `protoc-gen-twirp`, `mockgen` and `air`

## Everyday commands

```bash
make help               # list targets
make run                # DynamoDB Local + live-reloading API on :8000
make test               # unit tests with -race
make test-integration   # repository tests against DynamoDB Local
make test-e2e           # black-box tests; testcontainers manages containers
make test-all           # all of the above
make lint
make generate           # regenerate protobuf, Twirp and mocks
make build              # bin/user-settings
```

`make run` and the test targets source the matching file in `config/`:

| file | used by |
| --- | --- |
| `.env.local` | `make run` |
| `.env.test` | unit and integration tests |
| `.env.e2e` | e2e tests (endpoints are overridden by testcontainers at runtime) |

## Configuration

All configuration is environment variables, read once by `config.Load`.

| variable | default | purpose |
| --- | --- | --- |
| `HOST`, `PORT` | `0.0.0.0`, `8000` | listen address |
| `LOG_LEVEL`, `LOG_PRETTY` | `info`, `false` | slog level; text instead of JSON |
| `AWS_REGION` | required | region for the DynamoDB client |
| `DYNAMODB_ENDPOINT` | unset | set to use DynamoDB Local with static credentials |
| `USER_SETTINGS_TABLE_NAME` | required | table name |
| `USER_DIRECTORY_BASE_URL` | required | upstream user directory |
| `OTEL_SDK_DISABLED` | `false` | disable tracing |
| `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_TRACES_EXPORTER`, ... | | standard OpenTelemetry exporter settings |

## Local DynamoDB

`docker compose up` starts DynamoDB Local on `localhost:8900`, creates the
table from `docker/dynamodb/setup-entrypoint.sh`, and runs
[dynamodb-admin](https://github.com/aaronshaf/dynamodb-admin) on
`localhost:8001` so you can inspect items. Data persists under
`docker/dynamodb/data`, which is git-ignored.

## Test suites

**Unit** (`internal/...`): no I/O. Collaborators are gomock mocks generated
from the consumer-side interfaces. Run in well under a second.

**Integration** (`tests/integration`): the DynamoDB repository against a real
DynamoDB Local. Each test creates the table if needed and clears it after.
Needs `make docker-up`.

**End to end** (`tests/e2e`): builds the app through `app.New`, serves it on
an ephemeral port, and drives it over HTTP. `TestMain` starts DynamoDB Local
and a WireMock container for the user directory via testcontainers, so these
need Docker but no other setup. WireMock fixtures live in `tests/e2e/wiremock`.

## Changing the API

1. Edit `rpc/user_settings_api.proto` (or `rpc/session.proto`). Never reuse
   or renumber a field; add new fields with new numbers.
2. `make generate-proto`, then commit the output under `internal/generated`.
   CI fails if generated code is stale.
3. Implement the method on `api.Server`. The compile-time assertion
   `var _ proto.UserSettingsService = (*Server)(nil)` will tell you what is
   missing.
4. Add unit tests in `internal/api` for validation and error mapping, and an
   e2e test if the behaviour spans layers.

## Changing an interface

Interfaces that have mocks carry a `//go:generate mockgen ...` directive.
After editing one, run `make generate-mocks` and commit the mock alongside.

## Adding a setting definition

Edit `internal/prefsrepo/definitions/setting_definitions.go`. A definition
has a name, a type code (`s`, `b`, `n`), a `privacy` flag (whether it is
returned by `GetPrivacySettings`), and `defaults.global` plus an optional
`defaults.minor`. Add a case to `definition_test.go`.

## Adding an age rule

`regionalrulesrepo.GetRules` holds the rules in code today. Add an entry and a
test in `repository_test.go`. The rule value is a JSON document; see
`RuleAgeRestrictionValue`.

## Linting

`.golangci.yml` enables the standard set plus `errorlint`, `gocritic`,
`testifylint`, `bodyclose`, `noctx`, `nilerr`, `unparam`, `revive` and
`misspell`. `golangci-lint run --fix` handles most test-assertion findings.

## Release

`Dockerfile` produces a static binary in a distroless non-root image:

```bash
docker build -t user-settings .
docker run --rm -p 8000:8000 --env-file config/.env.local user-settings
```
