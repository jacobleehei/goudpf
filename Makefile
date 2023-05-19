# build linux and windows version of golang application
# save to bin folder
APP_NAME=goudpf

linux:
	GOOS=linux GOARCH=amd64 go build -o bin/linux/$(APP_NAME) main.go
windows:
	GOOS=windows GOARCH=amd64 go build -o bin/windows/$(APP_NAME).exe main.go
all: linux windows
