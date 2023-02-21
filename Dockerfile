FROM golang:1.18.3-stretch as builder

ARG VERSION

ENV VERSION=$VERSION
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -a -ldflags="-s -w" -o csvql ./

FROM debian:sid-slim

WORKDIR /app

COPY --from=builder ./app/csvql .