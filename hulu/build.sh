#!/bin/bash
#
##
HULU_VERSION="0.0.1"
GIT_COMMIT=$(git rev-parse HEAD)

go build -ldflags "-X main.version=$HULU_VERSION -X main.commit=$GIT_COMMIT" -o hulu

mkdir -p output

mv ./hulu output/
