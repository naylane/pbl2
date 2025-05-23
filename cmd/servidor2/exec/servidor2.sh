#!/bin/bash

chmod +x servidor2.sh

docker-compose build
docker-compose create
docker ps -a
docker-compose start servidor2
docker-compose logs -f 

# cd ./cmd/servidor2/exec
# ./servidor2.sh