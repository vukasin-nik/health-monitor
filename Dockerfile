# ---- Build stage ----
# If go.mod ends up requiring a newer Go than this tag, bump the tag.
FROM golang:1.27-alpine AS build
WORKDIR /src

# Copy dependency files first so Docker can cache the download layer.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/health-monitor .

# ---- Runtime stage ----
FROM alpine:3.20
RUN adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/health-monitor .
COPY targets.json .
USER app
EXPOSE 8080
ENTRYPOINT ["./health-monitor"]
