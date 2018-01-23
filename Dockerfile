FROM golang:1.9-stretch

COPY bashpush.go .
COPY run.sh .

RUN go get github.com/SlyMarbo/rss && \
    go get github.com/sideshow/apns2 && \
    go build bashpush.go

CMD ["/go/run.sh"]
