#!/bin/bash -xe

VERSION="v0.1"

ENVS=""
if [[ $# -gt 0 ]]; then
    if [ $1 == "--preprod" ]; then
        ENVS="preprod"
    elif [ $1 == "--live" ]; then
        ENVS="live"
    elif [ $1 == "--all" ]; then
        ENVS="all"
    fi
fi

if [ "$ENVS" = "preprod" ] || [ "$ENVS" = "all" ]; then
    ../push.sh eu-central-1 s3-web-client 979523904544 terraform Dockerfile "no" "$VERSION"
fi
if [ "$ENVS" = "live" ] || [ "$ENVS" = "all" ]; then
    ../push.sh eu-central-1 s3-web-client 643078788875 terraform Dockerfile "no" "$VERSION"
fi
