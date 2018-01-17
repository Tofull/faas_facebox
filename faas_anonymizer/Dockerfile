FROM golang:1.9.2-alpine as builder

ENTRYPOINT []

RUN apk --no-cache add make curl \
    && curl -sL https://github.com/openfaas/faas/releases/download/0.6.9/fwatchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog

RUN apk add --no-cache git mercurial \
    && go get github.com/machinebox/sdk-go/facebox \
    && apk del git mercurial

WORKDIR /go/src/faas_facebox

COPY handler.go .

RUN go install


###
FROM alpine:3.6

# Needed to reach the hub
RUN apk --no-cache add ca-certificates 

COPY --from=builder /usr/bin/fwatchdog  /usr/bin/fwatchdog
COPY --from=builder /go/bin/faas_facebox  /usr/bin/faas_facebox
ENV fprocess "/usr/bin/faas_facebox"

HEALTHCHECK --interval=5s CMD [ -e /tmp/.lock ] || exit 1
CMD ["/usr/bin/fwatchdog"]