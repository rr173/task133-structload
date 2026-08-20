# Official Go image with the full toolchain; frontend is native (no Node).
FROM golang:1.26.3

WORKDIR /app

# Go dependencies first for caching.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Pre-compile once; cache stays in the image. Does not affect model edits.
RUN go build ./...

# Container drops into a shell for interactive use by the evaluator.
CMD ["bash"]
