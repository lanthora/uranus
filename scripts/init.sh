#!/bin/bash
root=$(dirname $(cd $(dirname $0);pwd))

# 更新 Go 依赖
go mod tidy
