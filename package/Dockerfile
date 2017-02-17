FROM alpine:3.5
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/' /etc/apk/repositories
RUN apk add --update curl jq && rm -rf /var/cache/apk/*
COPY logging-helper entry.sh /usr/bin/
ENTRYPOINT ["/usr/bin/entry.sh"]
CMD ["logging-helper"]
