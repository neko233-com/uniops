# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build backend
RUN CGO_ENABLED=0 go build -o uniops ./cmd/uniops

# Build frontend
FROM node:22-alpine AS frontend-builder

WORKDIR /app/web

COPY web/package*.json ./
RUN npm ci

COPY web/ .
RUN npm run build

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/uniops .
COPY --from=frontend-builder /app/web/dist ./web/dist

EXPOSE 6020

CMD ["./uniops"]
