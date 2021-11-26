#!/bin/bash
pid=$(ps -ef|grep kimenet|grep -v grep|awk '{print $2}')
echo $pid
if [[ "$pid" != "" ]];then
  echo "发送信号"
  kill -9 $pid
fi

HULU_VERSION="0.0.1"
GIT_COMMIT=$(git rev-parse HEAD)

go build -ldflags "-X main.version=$HULU_VERSION -X main.commit=$GIT_COMMIT" -o kimenet

mkdir -p output

mv ./kimenet output/
