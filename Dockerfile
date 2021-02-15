FROM golang:latest AS builder
WORKDIR /go/src/github.com/xakepp35/StringRandomizerDemo
COPY go.mod .
COPY *.go .
RUN go build -v .
CMD ["./StringRandomizerDemo"]