FROM docker.arvancloud.ir/golang:latest AS builder
ENV HOME=/
ENV CGO_ENABLED=0
ENV GOOS=linux
WORKDIR /
COPY . .
ENV GOPROXY=http://registry.ik8s.ir/repository/golang.org/
RUN go get -d && go mod download && go build -a -ldflags "-s -w" -installsuffix cgo -o netlog .

FROM docker.arvancloud.ir/alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=builder /netlog .
CMD ["./netlog","--interface", "$INTERFACE","--format","json"]
