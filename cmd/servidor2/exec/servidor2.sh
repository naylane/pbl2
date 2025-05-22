#!/bin/bash

docker-compose build
docker-compose create
docker ps -a
docker-compose start servidor2
docker-compose logs -f 

# ./cmd/servidor2/exec/servidor2.sh