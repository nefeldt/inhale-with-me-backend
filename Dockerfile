# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.27-alpine AS build
WORKDIR /src

# Cache modules first.
COPY go.mod go.sum ./
RUN go mod download

# Build a fully static binary (pure-Go SQLite driver → no cgo needed).
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# ---- runtime stage ----
# distroless/static contains only CA certs + tzdata, no shell — small and safe.
FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=build /out/server /app/server

# The app creates DB_PATH's directory on startup; mount a persistent volume at /data.
ENV PORT=8080 \
    DB_PATH=/data/inhale.db \
    ENVIRONMENT=production

EXPOSE 8080
ENTRYPOINT ["/app/server"]
