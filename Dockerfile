FROM golang:alpine as builder

MAINTAINER PRIEWIENV "PRIEWIENV@users.noreply.github.com"

RUN apk --no-cache add git

WORKDIR $GOPATH/src/PRIEWIENV/BytomLit

ADD . $GOPATH/src/PRIEWIENV/BytomLit

RUN make

FROM alpine:latest as prod

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=0 $GOPATH/src/PRIEWIENV/BytomLit/node .

COPY --from=0 $GOPATH/src/PRIEWIENV/BytomLit/config.json .

EXPOSE 9000:9000

CMD ["./node --config.json"]











EXPOSE 9000:9000

ENTRYPOINT ["./node --config.json"]