FROM golang:1.21 AS builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=1 GOOS=linux go build -o tuna main.go

FROM alpine:latest
RUN apk add --no-cache sqlite-libs
WORKDIR /app
COPY --from=builder /app/tuna .
COPY backend/templates ./templates
COPY frontend/build ./static
CMD ["./tuna"]
