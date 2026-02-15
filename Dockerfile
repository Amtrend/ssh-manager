FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# First copy the dependency files so that Docker caches them.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Building a statically compiled binary
RUN CGO_ENABLED=0 GOOS=linux go build -o ssh-manager ./cmd/server

FROM alpine:latest

# Install certificates (needed for secure connections) and time zones
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Сopy only what is needed for work
COPY --from=builder /app/ssh-manager .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# Create a folder for the SQLite database
RUN mkdir ./data

EXPOSE 8080

CMD ["./ssh-manager"]
