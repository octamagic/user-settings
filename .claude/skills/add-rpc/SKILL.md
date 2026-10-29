---
name: add-rpc
description: Add a new RPC to the UserSettingsService end to end (proto, generated code, handler, service logic, tests). Use when asked to add an endpoint or RPC method.
argument-hint: <RpcName>
---

Add the RPC `$ARGUMENTS` to this service. Follow the steps in order and stop at any failing check.

1. Edit `rpc/user_settings_api.proto`: add the request and response messages and the rpc.
   New fields get new numbers. Never reuse or renumber a field; removed fields become `reserved`.
2. Run `make generate-proto`. It runs protoc twice on purpose (protoc-gen-twirp refuses mixed
   `go_package` values in one call); do not merge the invocations.
3. Implement the method on `api.Server` in `internal/api/settings.go`. The compile-time assertion
   `var _ proto.UserSettingsService = (*Server)(nil)` reports what is missing. The handler does
   validation and error mapping only: get the caller with `callerUrn`, use the nil-safe generated
   getters, map service errors with `errors.Is` against sentinels.
4. Put business rules in `prefsservice`. If the repository needs a new method, add it to the
   consumer-side interface in `prefsservice` and run `make generate-mocks`.
5. Tests, in the same change: a table-driven unit test in `internal/api` for validation and
   error mapping, a service test for the rule, and an e2e case in `tests/e2e` when the behaviour
   spans layers. Use `require` for errors and preconditions, `assert` for the rest.
6. Run `go vet ./... && golangci-lint run && make test`, then `make test-e2e`.
7. Commit the generated code with the proto change. Add the route
   `/twirp/usersettings.v1.UserSettingsService/$ARGUMENTS` to the RPC table in `README.md`
   and, if it changes a flow, to `docs/architecture.md`.
