FROM golang:1.10.3-alpine3.8 AS builder
LABEL stage=bot-go-builder
WORKDIR /go/src/github.com/alexkeating/open-source-monitor
COPY ./ ./
RUN apk --no-cache add ca-certificates
RUN go install github.com/alexkeating/open-source-monitor
FROM alpine:3.8
WORKDIR /go/bin
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/open-source-monitor /go/bin/open-source-monitor
CMD ["/go/bin/open-source-monitor"]
