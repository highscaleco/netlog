# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    libpcap-dev \
    gcc \
    musl-dev

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o netlog cmd/netlog/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    libpcap \
    ca-certificates

# Copy binary from builder
COPY --from=builder /app/netlog /usr/local/bin/netlog

# Set capabilities for packet capture
RUN setcap cap_net_raw,cap_net_admin=eip /usr/local/bin/netlog

# Create non-root user
RUN adduser -D -u 1000 netlog

# Switch to non-root user
USER netlog

# Set environment variables
ENV NETLOG_INTERFACE=eth0
ENV NETLOG_FORMAT=text

# Run the application
ENTRYPOINT ["netlog"] 