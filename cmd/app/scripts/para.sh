#!/bin/bash

docker-compose stop
docker-compose down --rmi all --volumes --remove-orphans
clear