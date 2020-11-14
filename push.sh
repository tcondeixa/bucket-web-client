#!/bin/bash -xe

VERSION=v0.2
ACCOUNT=tscondeixa
REPO=s3-web-client

docker build -f Dockerfile -t $ACCOUNT/$REPO:$VERSION  .
docker tag $ACCOUNT/$REPO:$VERSION $ACCOUNT/$REPO:latest
docker push $ACCOUNT/$REPO:$VERSION