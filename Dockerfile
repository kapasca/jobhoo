# --- Build stage ---
FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /jobhoo ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /jobhoo-seed ./cmd/seed

# --- Run stage ---
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /jobhoo ./jobhoo
COPY --from=build /jobhoo-seed ./jobhoo-seed
COPY --from=build /app/web ./web
EXPOSE 8080
ENTRYPOINT ["./jobhoo"]
