FROM docker.arvancloud.ir/golang:alpine AS builder
RUN apk add libpcap-dev gcc libc-dev
WORKDIR /
ENV CGO_ENABLED=1
ENV GOOS=linux
COPY . .
ENV GOPROXY=http://registry.ik8s.ir/repository/golang.org/
RUN go mod download && go build -o netlog cmd/netlog/main.go

FROM docker.arvancloud.ir/alpine:latest
RUN apk --no-cache add ca-certificates libpcap
COPY --from=builder /netlog /usr/local/bin/netlog
ENTRYPOINT ["netlog"]
