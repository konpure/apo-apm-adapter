FROM golang:1.21.5-bullseye AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o apo-apm-adapter ./cmd

FROM debian:bullseye-slim AS runner
WORKDIR /app
RUN apt-get update && apt-get -y install ca-certificates
COPY apm-adapter.yml /app/
COPY --from=builder /build/apo-apm-adapter /app/
CMD ["/app/apo-apm-adapter", "--config=/app/apm-adapter.yml"]