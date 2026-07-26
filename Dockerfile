# --- Build stage ---
FROM golang:1.22-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/seed ./cmd/seed

# --- Development stage ---
FROM golang:1.25-bookworm AS dev

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates git \
	&& rm -rf /var/lib/apt/lists/*

ENV GOBIN=/usr/local/bin
RUN go install github.com/air-verse/air@v1.61.7

COPY go.mod go.sum* ./
RUN go mod download

COPY .air.toml ./

EXPOSE 8070

CMD ["air", "-c", ".air.toml"]

# --- Runtime stage ---
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /out/server ./server
COPY --from=builder /out/seed ./seed
COPY migrations ./migrations
COPY web ./web

RUN mkdir -p uploads/resumes uploads/documents

ENV PORT=8080
ENV MIGRATIONS_DIR=/app/migrations
ENV TEMPLATES_DIR=/app/web/templates
ENV STATIC_DIR=/app/web/static
ENV UPLOAD_DIR=/app/uploads

EXPOSE 8080

CMD ["./server"]
