FROM alpine:3.6

RUN mkdir /app
WORKDIR /app

ADD detour-proxy /app/detour-proxy

RUN	apk --update add bash ca-certificates; \
    chmod +x /app/detour-proxy; \
    adduser -D -H -s /bin/false proxy_user proxy_user; \
    apk del bash;


EXPOSE 8080 1080 8888

USER proxy_user

ENTRYPOINT ["/app/detour-proxy"]
VOLUME ["/app/config.json"]
CMD ["--config", "/app/config.json"]