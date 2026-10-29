FROM golang:1.24.1-bullseye AS build
WORKDIR /go/src/app
COPY services/app .
RUN ./scripts/check-protos.sh
RUN apt update && apt install -y ca-certificates
RUN --mount=type=secret,id=git_credentials,target=/root/.git-credentials \
  --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  git config --global credential.helper store \
  && go env -w GOPRIVATE="github.com/soundcloud/*" \
  && go mod download \
  && CGO_ENABLED=0 GOOS=linux go build -o ./bin/user-settings ./cmd/api

FROM scratch

WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /go/src/app/bin/user-settings .
CMD ["/app/user-settings"]
