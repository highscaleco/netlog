FROM docker.arvancloud.ir/golang:alpine AS builder
RUN apk add libpcap-dev gcc libc-dev
ENV HOME=/
ENV CGO_ENABLED=0
ENV GOOS=linux
WORKDIR /
COPY . .
ENV GOPROXY=http://registry.ik8s.ir/repository/golang.org/
RUN go get -d && go mod download && go build -a -ldflags "-s -w" -installsuffix cgo -o netlog ./cmd/netlog

FROM docker.arvancloud.ir/alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=builder /netlog .
CMD ["./netlog","--interface", "$INTERFACE","--format","json"]
