#!/bin/bash


cd /go/src/goagent/

if [[ "$1" == "consumer" ]]; then
  echo "Starting consumer agent..."
  GODEBUG=gctrace=1 ./main -config=config/consumer.yaml 2>/root/logs/gogc.log
#  GODEBUG=gctrace=1 ./main -config=config/consumer.yaml
elif [[ "$1" == "provider-small" ]]; then
  echo "Starting small provider agent..."
  GODEBUG=gctrace=1 ./main -config=config/provider-s.yaml 2>/root/logs/gogc.log
#   GODEBUG=gctrace=1 ./main -config=config/provider-s.yaml
elif [[ "$1" == "provider-medium" ]]; then
  echo "Starting medium provider agent..."
  GODEBUG=gctrace=1 ./main -config=config/provider-m.yaml 2>/root/logs/gogc.log
elif [[ "$1" == "provider-large" ]]; then
  echo "Starting large provider agent..."
  GODEBUG=gctrace=1 ./main -config=config/provider-l.yaml 2>/root/logs/gogc.log
else
  echo "Unrecognized arguments, exit."
  exit 1
fi


