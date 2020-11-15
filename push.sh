#!/bin/bash -xe

VERSION=v0.4
ACCOUNT=tscondeixa
REPO=s3-web-client

docker build -f Dockerfile -t $ACCOUNT/$REPO:$VERSION  .
docker tag $ACCOUNT/$REPO:$VERSION $ACCOUNT/$REPO:latest
docker push $ACCOUNT/$REPO:$VERSION