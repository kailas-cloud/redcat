# ---------------------------
# Stage 1: build
# ---------------------------
FROM golang:1.24 AS builder

WORKDIR /app/redcat

# сначала только мод-файлы (кеш)
COPY redcat/go.mod ./
COPY redcat/go.sum ./
RUN go mod tidy && go mod download

COPY redcat/. .


RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /build/redcat ./cmd/redcat

# ---------------------------
# Stage 2: runtime
# ---------------------------
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

COPY --from=builder /build/redcat /redcat

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/redcat"]
