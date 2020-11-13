#!/bin/bash -xe

VERSION=v0.1
ACCOUNT=tscondeixa
REPO=s3-web-client

docker build -f Dockerfile -t $ACCOUNT/$REPO:$VERSION  .
docker push $ACCOUNT/$REPO:$VERSION