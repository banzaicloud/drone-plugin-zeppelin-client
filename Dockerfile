FROM alpine:3.2
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
ADD zeppelin-client /bin/
ENTRYPOINT ["/bin/zeppelin-client"]