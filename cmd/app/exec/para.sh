#!/bin/bash

# cd ./cmd/app/exec
#chmod +x para.sh
# ./para.sh

docker-compose stop
docker-compose down --rmi all --volumes --remove-orphans
clear