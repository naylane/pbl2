#!/bin/bash

chmod +x servidor3.sh

docker-compose build
docker-compose create
docker ps -a
docker-compose start servidor3
docker-compose logs -f 

# cd ./cmd/servidor3/exec
# ./servidor3.sh