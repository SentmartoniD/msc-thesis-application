# syntax=docker/dockerfile:1

# ---------- build ----------
FROM golang:1.25-alpine AS build

WORKDIR /src

# Dependencies first: this layer stays cached unless go.mod/go.sum change,
# so ordinary source edits skip the download entirely.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG BUILD_VERSION=dev
ARG BUILD_COMMIT=unknown
ARG TARGETOS=linux
ARG TARGETARCH=amd64

# Builds every cmd/ package into /out/, named after its directory:
#   ./cmd/api     -> /out/api
#   ./cmd/migrate -> /out/migrate
#
# CGO_ENABLED=0 produces static binaries, so the runtime stage needs no libc.
# -trimpath strips absolute build paths, making the build reproducible: the
# same source yields the same image digest, which matters when an experiment
# run has to be traceable to an exact artefact.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/ ./cmd/...

# ---------- runtime ----------
FROM gcr.io/distroless/static-debian12:nonroot

ARG BUILD_VERSION=dev
ARG BUILD_COMMIT=unknown

# config.Load reads these from the environment; baking them in means every log
# line identifies the image that produced it.
ENV BUILD_VERSION=${BUILD_VERSION} \
    BUILD_COMMIT=${BUILD_COMMIT}

COPY --from=build /out/ /app/

EXPOSE 8000 9090

USER nonroot:nonroot

# Default role. The migration Job overrides this with command: ["/app/migrate"].
ENTRYPOINT ["/app/api"]