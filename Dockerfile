FROM golang:1.20 as builder

ENV GO111MODULE=on

# Working directory
WORKDIR /build

COPY . .

# Build app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./app

# final stage
FROM alpine:3.14

# Copy binary from builder
COPY --from=builder /build /

# Run server command
ENV TZ Asia/Saigon

# expose some necessary port
EXPOSE 8080
ENTRYPOINT ["/app", "start", "--env", ".env", "--config", "config.prd.toml"]