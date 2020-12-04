version := $(shell /bin/date "+%Y-%m-%d %H:%M")

build:
	go build -ldflags="-s -w" -ldflags="-X 'main.BuildTime=$(version)'" ezioctl.go
	$(if $(shell command -v upx), upx ezioctl)
mac:
	GOOS=darwin go build -ldflags="-s -w" -ldflags="-X 'main.BuildTime=$(version)'" -o ezioctl-darwin ezioctl.go
	$(if $(shell command -v upx), upx ezioctl-darwin)
win:
	GOOS=windows go build -ldflags="-s -w" -ldflags="-X 'main.BuildTime=$(version)'" -o ezioctl.exe ezioctl.go
	$(if $(shell command -v upx), upx ezioctl.exe)
linux:
	GOOS=linux go build -ldflags="-s -w" -ldflags="-X 'main.BuildTime=$(version)'" -o ezioctl-linux ezioctl.go
	$(if $(shell command -v upx), upx ezioctl-linux)
