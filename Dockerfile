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

# =========================================================================== #

# # Ultra lightweight Dockerfile

# FROM golang:1.22-alpine as builder

# # Install minimal dependencies
# RUN apk add --no-cache git ca-certificates

# WORKDIR /build
# COPY . .

# # Ultra lightweight build
# RUN CGO_ENABLED=0 GOOS=linux go build \
#     -a -installsuffix cgo \
#     -ldflags='-w -s -extldflags "-static"' \
#     -trimpath \
#     -o trading main.go

# # Ultra minimal final stage
# FROM scratch

# # Copy only certificates and binary
# COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# COPY --from=builder /build/trading /trading
# COPY --from=builder /build/config.lightweight.toml /config.toml

# EXPOSE 8080

# ENTRYPOINT ["/trading", "start", "--config", "/config.toml"]