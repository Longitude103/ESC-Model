# syntax=docker/dockerfile:1

## BUILD
FROM golang:1.21 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o bin/wwum2020-amd64-linux

## Deploy
FROM clearlinux:latest

WORKDIR /app

COPY --from=build /app/bin/wwum2020-amd64-linux ./wwum2020-amd64-linux
COPY .env .
#VOLUME /home/heath/Documents/code/wwum2020/bin ./bin

#ENTRYPOINT ["./app", "bash"]