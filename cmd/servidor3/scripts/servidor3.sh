#!/bin/bash

docker-compose build
docker-compose create
docker ps -a
docker-compose start servidor3
docker-compose logs -f 