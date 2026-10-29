# user-settings

A small, production-shaped Go service for storing per-user settings and
answering privacy questions about users, such as whether they are a minor
under the age rules of their country.

It was originally built for a music streaming platform's backend. This copy
has been scrubbed of that company's infrastructure and internal libraries,
and reworked in a series of focused commits. The git history is meant to be
read: each commit after the initial import fixes or improves one thing and
says why.

## What it does

The service exposes a [Twirp](https://github.com/twitchtv/twirp) RPC API
(JSON or protobuf over HTTP) with four methods:

| RPC | Purpose |
| --- | --- |
| `GetSettings` | Return the caller's settings, optionally filtered by name |
| `PutSetting` | Create or overwrite one setting for the caller |
| `GetPrivacySettings` | Return the privacy-relevant settings of other users, filling in defaults for anything unset |
| `FetchUsersPrivacyAttributes` | Report whether each user is a minor under their country's age rules |

Settings are stored in a single DynamoDB table keyed by user URN and setting
name. Ages come from an upstream user directory over HTTP. Age thresholds
per country live in an in-memory rules repository.

## Quick start

Prerequisites: Go 1.24+, Docker, `protoc` (only if you change the protos).

```bash
make docker-up              # DynamoDB Local on :8900, admin UI on :8001
set -a; . ./config/.env.local; set +a
go run ./cmd/api
```

Then, in another shell:

```bash
B=http://localhost:8000/twirp/usersettings.v1.UserSettingsService

curl -s -X POST -H 'Content-Type: application/json' $B/PutSetting \
  -d '{"user_session":{"user_urn":"urn:user:42"},"setting":{"name":"volume","type":"NUMBER","value":"80"}}'
# {"setting":{"user_urn":"urn:user:42","name":"volume","type":"NUMBER","value":"80","updated_at":"1793295322"}}

curl -s -X POST -H 'Content-Type: application/json' $B/GetSettings \
  -d '{"user_session":{"user_urn":"urn:user:42"},"setting_names":["volume"]}'
# {"settings":[{"user_urn":"urn:user:42","name":"volume","type":"NUMBER","value":"80","updated_at":"1793295322"}]}

curl -s -X POST -H 'Content-Type: application/json' $B/PutSetting \
  -d '{"user_session":{"user_urn":"urn:user:42"},"setting":{"name":"theme","type":"STRING","value":"dark"}}'
# {"code":"invalid_argument","msg":"setting.value value must be a valid JSON document", ...}
```

Setting values are JSON documents, so a string value is `"\"dark\""`.

## Tests

```bash
make test               # unit tests, no I/O
make test-integration   # repository against DynamoDB Local (make docker-up first)
make test-e2e           # full service over HTTP; testcontainers starts DynamoDB and WireMock
make lint
```

See [docs/development.md](docs/development.md) for the full workflow and
[docs/architecture.md](docs/architecture.md) for how the code is organised and
why.

## Layout

```
cmd/api/                 entry point: load config, build the app, run the server
internal/
  app/                   composition root: wires clients, repos, services, handler
  api/                   Twirp handlers; maps domain errors to Twirp codes
  httpserver/            router, middleware, graceful shutdown
  prefsservice/          settings use cases
    domain/              minor-protection decision (age parsing, rules)
  prefsrepo/             DynamoDB repository and setting definitions
  rulesservice/          applies per-country age rules
  regionalrulesrepo/     in-memory rule store
  userdirectory/         HTTP client for the upstream user directory
  mappers/               proto <-> domain conversion
  config/ logging/ telemetry/ awsdynamo/
  generated/             protoc + Twirp output (committed; `make generate`)
rpc/                     .proto contracts
tests/integration/       repository tests against DynamoDB Local
tests/e2e/               black-box tests with testcontainers
```

## Infrastructure and delivery

`terraform/` runs the service on AWS (ECS Fargate behind an internal ALB, DynamoDB,
ECR, CloudWatch, autoscaling, a GitHub OIDC deploy role) and is composed entirely
from [tf-tools](https://github.com/octamagic/tf-tools) modules; the two pieces
it needed that the library lacked were contributed back rather than copied in. The
GitHub workflows are assembled from
[actions-toolkit](https://github.com/octamagic/actions-toolkit) reusable
workflows: Go CI, image build and scan, Terraform plan-on-PR / apply-on-merge, and an
OIDC-authenticated ECS deploy. Details in [terraform/README.md](terraform/README.md).

## Status and known gaps

- Rules are hard-coded (GB and US, 18+). A real deployment would load them
  from configuration or a table.
- Only one setting definition ships. Definitions are the place where a
  setting's type, privacy flag and defaults live; see `internal/prefsrepo/definitions`.
- `GetSettings` queries a user's whole partition and filters by name in
  memory. Users have a handful of settings, so this is one cheap query rather
  than a `BatchGetItem`; revisit if the per-user setting count grows.
- There is no authentication. The `user_session` in each request is trusted,
  as it was behind the original platform's API gateway.
- Terraform pins tf-tools `v0.3.0`, which is cut once the companion tf-tools branch
  (DynamoDB module, `track_latest_task_definition`) merges. Until then `make tf-test`
  validates against a local checkout.
