FROM golang:1.18.3-stretch as builder

ARG VERSION

ENV VERSION=$VERSION
ENV GOOS=linux
ENV CGO_ENABLED=0

WORKDIR /app

ADD go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags="-s -w" -o csvql ./

FROM debian:sid-slim

WORKDIR /app

COPY --from=builder ./app/csvql .