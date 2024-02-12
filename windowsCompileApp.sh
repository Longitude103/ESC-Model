# !/bin/bash

echo "What is the version number for this release (don't include 'v')?"
read VERSION

GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC="x86_64-w64-mingw32-gcc" go build -o bin/wwumm-amd64-win_v$VERSION.exe

echo "Done Compiling windows wwumm-amd64-win_v${VERSION}, look in bin/"
