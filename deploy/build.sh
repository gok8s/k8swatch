#!/usr/bin/env bash
export GO111MODULE=on 
export GOPROXY=https://goproxy.io
projectName=k8swatch
tag=0.26.9
registryAddr=""

GOOS=linux GOARCH=amd64 go build -o bin/$projectName  github.com/gok8s/k8swatch
chmod +x bin/$projectName
docker build -t $registryAddr/cre/$projectName:$tag .
docker push $registryAddr/cre/$projectName:$tag
