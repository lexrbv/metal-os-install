OUT := metal-os-install
PKG := github.com/lexrbv/metal-os-install
VERSION := $(shell git describe --always --long --dirty)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)

all: run

vet:
	@go vet ${PKG_LIST}

build: vet
	go build -ldflags "-s -w -X '${PKG}/build.Version=${VERSION}'" -o ./metal-os-install main.go

run: server
	./${OUT}

clean:
	-@rm ${OUT} ${OUT}-v*

.PHONY: run server build vet
