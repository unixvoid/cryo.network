FROM alpine

RUN apk add --update ca-certificates
COPY cryon /
COPY config.gcfg /

CMD ["/cryon"]
