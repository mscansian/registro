#!/bin/sh
echo Building numercfd/registro:build
docker build -t numercfd/registro:build . -f Dockerfile.build
docker create --name extract numercfd/registro:build
docker cp extract:/go/src/github.com/numercfd/registro/registro ./registro
docker rm -f extract

echo Building numercfd/registro:latest
docker build --no-cache -t numercfd/registro:latest .
rm ./registro
