# syntax=docker/dockerfile:1

## BUILD
# go-sqlite3 needs CGO, so build on the full Debian-based Go image.
FROM golang:1.26-trixie AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o bin/esc-model-amd64-linux

## RUN
# Same Debian release as the build image so the glibc versions match.
FROM debian:trixie-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=build /app/bin/esc-model-amd64-linux ./esc-model

# The database login is not baked into the image. Either run
#   docker run -it <image> ./esc-model init -u <username> -p <password>
# inside the container, or mount an existing .configjson to /app/.configjson.
ENTRYPOINT ["./esc-model"]
