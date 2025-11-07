# ---------- build stage ----------
FROM golang:1.24 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/storage ./cmd/storage

# ---------- final stage ----------
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /bin/storage /app/storage

EXPOSE 9001
ENTRYPOINT ["/bin/sh", "-c", "/app/storage"]
