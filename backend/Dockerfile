# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder
WORKDIR /src

# Cache deps separately
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /server ./cmd/server

# Runtime: distroless static (no shell, no package manager)
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /server /server
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
