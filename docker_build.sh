#!/bin/bash

CGO_ENABLED=0 go build -a -installsuffix cgo
sudo docker build -t="wangjia184/detour-proxy:latest" .
sudo docker push wangjia184/detour-proxy:latest
