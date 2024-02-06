#!/usr/bin/env zsh

echo "What is the version number for this release (don't include 'v')?"

read VERSION

#GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=mingw64-configure go build -o bin/wwumm-amd64.exe
GOOS=darwin GOARCH=amd64 go build -o bin/wwumm-amd64-darwin_v$VERSION
GOOS=darwin GOARCH=arm64 go build -o bin/wwumm-arm64-darwin_v$VERSION
GOOS=linux GOARCH=amd64 go build -o bin/wwumm-amd64-linux_v$VERSION

echo "Done Compiling wwumm_v${VERSION}, look in bin/"
