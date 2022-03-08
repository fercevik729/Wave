FROM golang:1.17-alpine

COPY . /app
WORKDIR /app

RUN go mod download

RUN go build -o /wave
