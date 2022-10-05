# syntax=docker/dockerfile:1
FROM golang:1.18-alpine

WORKDIR /server

COPY go.mod ./
COPY go.sum ./

RUN go mod download
COPY ./ ./
RUN go build main.go 
EXPOSE 9090

CMD [ "./main" ]

