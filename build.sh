#!/bin/bash

GOOS=windows GOARCH=amd64 go build -o ./build/vozdns.exe
GOOS=darwin GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-s -w -extldflags=-static" -o ./build/vozdns_amd64_darwin
GOOS=darwin GOARCH=arm64 go build -a -installsuffix cgo -ldflags="-s -w -extldflags=-static" -o ./build/vozdns_arm64_darwin
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-s -w -extldflags=-static" -o ./build/vozdns_amd64
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -installsuffix cgo -ldflags="-s -w -extldflags=-static" -o ./build/vozdns_arm64