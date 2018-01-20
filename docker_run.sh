#!/bin/bash

IMAGE_URI=wangjia184/detour-proxy:latest

#sudo docker pull $IMAGE_URI
sudo docker rm -f detour-proxy
sudo docker run -ti --name detour-proxy \
	-p 1080:1080 \
	$IMAGE_URI --role client --brokerUrl wss://www.013201.cn/api/stream/
	