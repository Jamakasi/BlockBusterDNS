#!/bin/sh
docker buildx build  --no-cache --platform amd64 --output type=local,dest=out/ ../docker/
mv out/dns ./release
rm -rf out
