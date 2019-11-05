#!/usr/bin/env bash
set -e

rm -rf /tmp/test-atreus
mkdir /tmp/test-atreus
atreus new a -d /tmp/test-atreus
cd /tmp/test-atreus/a/cmd && go build
if [ $? -ne 0 ]
then
  echo "Failed: all"
  exit 1
fi
atreus new b -d /tmp/test-atreus --grpc
cd /tmp/test-atreus/b/cmd && go build
if [ $? -ne 0 ]
then
  echo "Failed: --grpc"
  exit 1
fi
atreus new c -d /tmp/test-atreus --http
cd /tmp/test-atreus/c/cmd && go build
if [ $? -ne 0 ]
then
  echo "Failed: --http"
  exit 1
fi
