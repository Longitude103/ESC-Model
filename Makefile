version = 1.6.15

build-win:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc go build -o bin/wwumm-amd64-win_v$(version).exe

build:
	GOOS=darwin GOARCH=amd64 go build -o bin/wwumm-amd64-darwin_v$(version)
	GOOS=darwin GOARCH=arm64 go build -o bin/wwumm-arm64-darwin_v$(version)
	GOOS=linux GOARCH=amd64 go build -o bin/wwumm-amd64-linux_v$(version)
