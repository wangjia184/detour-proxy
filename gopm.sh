#!/bin/bash

$GOPATH/bin/gopm gen
$GOPATH/bin/gopm get -g
#go get -u github.com/golang/protobuf/protoc-gen-go
cat ./.gopmfile
