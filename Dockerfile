FROM golang:1.13rc1-alpine3.10 as builder

MAINTAINER PRIEWIENV "PRIEWIENV@users.noreply.github.com"

ARG http_proxy

ARG https_proxy=$http_proxy

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories \
    && apk --no-cache add git make

WORKDIR /go/src/BytomLit

ADD . /go/src/BytomLit

RUN make

FROM alpine:3.10 as prod

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories \
    && apk --no-cache add ca-certificates

WORKDIR /var/bytomlit/

COPY --from=0 /go/src/BytomLit/node .

COPY --from=0 /go/src/BytomLit/config.json .

EXPOSE 9000

CMD ["./node", "config.json"]
