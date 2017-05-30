FROM golang:1.8

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

ADD main.go /go/src/app/

RUN go-wrapper download
RUN go-wrapper install

CMD ["app"]
