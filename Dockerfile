FROM alpine:3.4
MAINTAINER Daisuke Fujita <dtanshi45@gmail.com> (@dtan4)

ENV DOCKER_COMPOSE_VERSION 1.6.0
ENV GLIBC_VERSION 2.23-r3

RUN apk --no-cache --update add docker openssl && \
    wget -O /usr/local/bin/docker-compose https://github.com/docker/compose/releases/download/$DOCKER_COMPOSE_VERSION/docker-compose-Linux-x86_64 && \
    chmod +x /usr/local/bin/docker-compose && \
    wget -O glibc.apk https://github.com/andyshinn/alpine-pkg-glibc/releases/download/$GLIBC_VERSION/glibc-$GLIBC_VERSION.apk && \
    wget -O glibc-bin.apk https://github.com/andyshinn/alpine-pkg-glibc/releases/download/$GLIBC_VERSION/glibc-bin-$GLIBC_VERSION.apk && \
    apk add --allow-untrusted glibc.apk glibc-bin.apk && \
    /usr/glibc-compat/sbin/ldconfig /usr/glibc-compat/lib && \
    rm glibc.apk glibc-bin.apk

ENV GOPATH /go
COPY . /go/src/github.com/dtan4/paus
RUN apk --no-cache --update add curl git go make mercurial && \
    cd /go/src/github.com/dtan4/paus && \
    make deps && \
    make && \
    mkdir /app && \
    cp bin/paus-frontend /app/paus-frontend && \
    cp -R templates /app/templates && \
    cd /app && \
    rm -rf /go && \
    apk del --purge curl git go make mercurial

WORKDIR /app
EXPOSE 8080

CMD ["./paus-frontend"]
