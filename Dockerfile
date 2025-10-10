FROM golang:1.24.2-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git ca-certificates tzdata
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o hotspot ./cmd/app


FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/hotspot /usr/local/bin/hotspot
COPY --from=builder /app/configs /app/configs
EXPOSE 8080
CMD ["hotspot", "serve"]
