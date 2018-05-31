#!/bin/bash

sudo docker tag `$1` registry.cn-hangzhou.aliyuncs.com/gaosai/tinchi.etcd:latest

sudo docker push registry.cn-hangzhou.aliyuncs.com/gaosai/tinchi.etcd:latest