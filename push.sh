#!/usr/bin/env bash
# 编译
#CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/linux
#CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/window64.exe
#go build -o dist/macos
# 上传到git仓库
git add .
git commit -m "优化配置管理部分和docker"
git push -f origin master
