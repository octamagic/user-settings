FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /out/user-settings ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/user-settings /user-settings
EXPOSE 8000
ENTRYPOINT ["/user-settings"]
