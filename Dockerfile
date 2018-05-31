FROM golang:1.10.2 as builder
WORKDIR /
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o proxy main.go

FROM alpine:latest
WORKDIR /
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=builder /proxy .
ENV AF_PROXY_PORT=4001
EXPOSE 4001/tcp
CMD ["./proxy"]