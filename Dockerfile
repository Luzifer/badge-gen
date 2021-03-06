FROM golang:alpine as builder

COPY . /go/src/github.com/Luzifer/badge-gen
WORKDIR /go/src/github.com/Luzifer/badge-gen

RUN set -ex \
 && apk add --update git \
 && go install \
      -ldflags "-X main.version=$(git describe --tags --always || echo dev)" \
      -mod=readonly

FROM alpine:latest

LABEL maintainer "Knut Ahlers <knut@ahlers.me>"

RUN set -ex \
 && apk --no-cache add \
      ca-certificates

COPY --from=builder /go/bin/badge-gen /usr/local/bin/badge-gen

EXPOSE 3000
VOLUME ["/config"]

ENTRYPOINT ["/usr/local/bin/badge-gen"]
CMD ["--config", "/config/config.yaml"]

# vim: set ft=Dockerfile:
