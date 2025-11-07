# ---------- build stage ----------
FROM golang:1.24 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/distributor ./cmd/distributor

# ---------- final stage ----------
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /bin/distributor /app/distributor

EXPOSE 8080 9000
ENTRYPOINT ["/bin/sh", "-c", "/app/distributor"]
