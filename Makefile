all: build

deps:
	govendor fetch +missing

build:
	env GOOS=linux GOARCH=arm GOARM=5 go build -v .
