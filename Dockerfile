FROM gliderlabs/alpine:3.1

MAINTAINER Knut Ahlers <knut@ahlers.me>

RUN apk --update add wget && \
    wget --no-check-certificate https://gobuilder.me/get/github.com/Luzifer/badge-gen/badge-gen_master_linux-386.zip && \
    unzip badge-gen_master_linux-386.zip

ENV PORT 3000
EXPOSE 3000
ENTRYPOINT ["/badge-gen/badge-gen"]
