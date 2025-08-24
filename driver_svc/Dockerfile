# Build
FROM golang:1.24.6-alpine AS builder
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOTOOLCHAIN=auto
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/driver_svc ./cmd/driver_svc

# Runtime
FROM alpine:3.19
RUN adduser -D -H app && apk add --no-cache ca-certificates
USER app
WORKDIR /app
COPY --from=builder /out/driver_svc /app/driver_svc
EXPOSE 8081
ENTRYPOINT ["./driver_svc"]