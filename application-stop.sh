#!/bin/sh

#起動コンテナを停止
docker stop $(docker ps -q)
