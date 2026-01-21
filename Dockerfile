FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /redcat ./cmd/redcat

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /redcat /redcat

EXPOSE 8080

ENTRYPOINT ["/redcat"]
