FROM golang:1.13.7-buster

WORKDIR /go/src/app
COPY . .

RUN go get -d -v
RUN go build -v

CMD ["./app"]
