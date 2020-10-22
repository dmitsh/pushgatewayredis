FROM alpine

RUN apk update && \
    apk upgrade && \
    rm -rf /var/cache/apk/*

COPY pushgateway /usr/local/bin/pushgateway

USER nobody

ENTRYPOINT /usr/local/bin/pushgateway
