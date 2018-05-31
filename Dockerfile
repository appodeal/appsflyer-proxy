FROM golang:1.10.2 as builder
WORKDIR /
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o proxy main.go

FROM alpine:latest
WORKDIR /
COPY --from=builder /proxy .
CMD ["./proxy"]