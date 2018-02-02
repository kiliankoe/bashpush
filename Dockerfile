FROM golang:1.9-alpine as build

COPY bashpush.go .

RUN apk update && apk add git

RUN go get github.com/SlyMarbo/rss && \
    go get github.com/sideshow/apns2 && \
    go build bashpush.go

FROM alpine:latest

COPY --from=build /go/bashpush /bashpush/
COPY run.sh /bashpush/

CMD ["/bashpush/run.sh"]
