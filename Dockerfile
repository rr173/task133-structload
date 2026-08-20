# Build-time image with the exact Go toolchain version.
FROM docker.m.daocloud.io/library/golang:1.26.3-bookworm AS build

WORKDIR /src

# Copy dependency manifests first to leverage build cache.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOTOOLCHAIN=local go build -o /out/structload .

# Runtime image: minimal, CGO-free binary on alpine.
FROM docker.m.daocloud.io/library/alpine:3.20

COPY --from=build /out/structload /usr/local/bin/structload

WORKDIR /data
ENV GOTOOLCHAIN=local

# Default: run the self-check and exit. Override with --addr=:8080 for serving.
ENTRYPOINT ["structload"]
CMD ["--smoke-test", "--db=/data/structload.db"]
