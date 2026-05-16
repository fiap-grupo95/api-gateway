FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /api-gateway ./main.go

FROM alpine:3.21
RUN adduser -D -u 1000 appuser
COPY --from=builder /api-gateway /api-gateway
USER appuser
EXPOSE 8080
ENTRYPOINT ["/api-gateway"]
