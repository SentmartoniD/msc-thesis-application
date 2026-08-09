FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG BUILD_VERSION=dev
ARG BUILD_COMMIT=unknown
ARG TARGETOS=linux
ARG TARGETARCH=amd64

# CGO_ENABLED=0 produces a static binary, so the runtime stage needs no libc.
# -trimpath strips absolute build paths, making the build reproducible: the
# same source yields the same image digest, which matters when an experiment
# run has to be traceable to an exact artefact.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot

ARG BUILD_VERSION=dev
ARG BUILD_COMMIT=unknown

# config.Load reads these from the environment; baking them in means every log
# line identifies the image that produced it.
ENV BUILD_VERSION=${BUILD_VERSION} \
    BUILD_COMMIT=${BUILD_COMMIT}

COPY --from=build /out/api /app/api

EXPOSE 8000 9090

USER nonroot:nonroot
ENTRYPOINT ["/app/api"]