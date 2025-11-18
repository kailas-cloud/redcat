# ---------------------------
# Stage 1: build
# ---------------------------
FROM golang:1.22 AS builder

WORKDIR /app

# Копируем весь модуль разом
COPY redcat ./redcat

WORKDIR /app/redcat

# Моды и билд
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /app/redcat ./cmd/redcat

# ---------------------------
# Stage 2: runtime
# ---------------------------
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

COPY --from=builder /app/redcat /app/redcat

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/app/redcat"]
