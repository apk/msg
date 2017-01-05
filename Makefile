IMAGE_NAME=apky/msgsrv
VERSION=$(shell git describe)

msgsrv: msgsrv.go
	@echo Compiling...
	CGO_ENABLED=0 go build msgsrv.go

build: msgsrv
	@echo Building \"${IMAGE_NAME}\"...
	docker build -t=${IMAGE_NAME} .
	@echo ok

release: build
	@echo Tagging and Pushing \"${IMAGE_NAME}\":${VERSION}...
	docker tag -f ${IMAGE_NAME} ${IMAGE_NAME}:${VERSION}
	docker push ${IMAGE_NAME}:${VERSION}
	docker push ${IMAGE_NAME}:latest

.PHONY: build release
