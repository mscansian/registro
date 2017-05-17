#!/bin/sh
echo Building mscansian/registro:build
docker build -t mscansian/registro:build . -f Dockerfile.build
docker create --name extract mscansian/registro:build
docker cp extract:/go/src/github.com/mscansian/registro/registro ./registro
docker rm -f extract

echo Building mscansian/registro:latest
docker build --no-cache -t mscansian/registro:latest .
rm ./registro
