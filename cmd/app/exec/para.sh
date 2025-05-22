#!/bin/bash

chmod +x para.sh

docker-compose stop
docker-compose down --rmi all --volumes --remove-orphans
clear

# cd ./cmd/app/exec
# ./para.sh