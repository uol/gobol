FROM alpine:latest

ARG zkIP

RUN apk update && apk upgrade && apk add bash

COPY election /usr/bin/

CMD election