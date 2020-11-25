#!/bin/bash -xe

VERSION=v0.7
ACCOUNT=tscondeixa
REPO=bucket-web-client

docker login --username tscondeixa --password $(cat ~/.dockerhub)
docker build -f Dockerfile -t $ACCOUNT/$REPO:$VERSION  .
docker tag $ACCOUNT/$REPO:$VERSION $ACCOUNT/$REPO:latest
docker push $ACCOUNT/$REPO:$VERSION