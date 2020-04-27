#!/bin/bash

# Если под Windows, добавляем команду sudo
# if [[ "$OSTYPE" == "msys" ]]; then alias sudo=""; fi



echo "гасим docker-compose"
docker-compose down


# компилируем. линкуем статически под линукс
# env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a  .

# ключи нужны для компиляции sqlite3 при CGO_ENABLED=1
#echo "Кросскомпиляция на компьютере разработчика"
#env CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' .

echo "Кросскомпиляция в докере. Сделано чтобы компилировать под windows. 1.14.2 версия go на момент написания кода."
# docker run --rm -it -v "$PWD":/usr/src/myapp -w /usr/src/myapp -e CGO_ENABLED=1 -e GOOS=linux golang:1.14.2 go build -a -ldflags '-linkmode external -extldflags "-static"'
docker run --rm -it -v "$PWD":/usr/src/myapp -w /usr/src/myapp -e CGO_ENABLED=0 -e GOOS=linux golang:1.14.2 go build -a 


echo "build a docker image"
docker build -t vadimivlev/log-generator:latest . 

echo "push the docker image" 
docker login
docker push vadimivlev/log-generator:latest

