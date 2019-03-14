#!/bin/bash

export GOOS=linux; go build .
cp kubepoc ./artifacts/simple-image/kubepoc-apiserver
docker build -t kubepoc-apiserver:latest ./artifacts/simple-image
