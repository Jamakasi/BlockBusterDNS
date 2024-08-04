#!/bin/sh
docker buildx build  --no-cache --platform amd64 --output=type=docker -t chr-x86_64 ../docker/
docker save chr-x86_64 > chr-x86_64.tar
docker image rm chr-x86_64