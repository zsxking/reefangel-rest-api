all: deps build

deps:
	glide -q install

build:
	env GOOS=linux GOARCH=arm GOARM=5 go build -v .
