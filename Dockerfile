FROM golang:alpine

MAINTAINER Knut Ahlers <knut@ahlers.me>

ADD . /go/src/github.com/Luzifer/badge-gen
WORKDIR /go/src/github.com/Luzifer/badge-gen

RUN set -ex \
 && apk add --update git ca-certificates \
 && go install -ldflags "-X main.version=$(git describe --tags || git rev-parse --short HEAD || echo dev)" \
 && apk del --purge git

EXPOSE 3000

VOLUME ["/config"]

ENTRYPOINT ["/go/bin/badge-gen"]
CMD ["--config", "/config/config.yaml"]
