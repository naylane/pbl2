#!/bin/bash

docker-compose build
docker-compose create
docker ps -a
docker-compose start servidor1 
docker-compose logs -f
#Para o veiculo:
#docker-compose start veiculo
#docker exec -it veiculo sh
#./veiculo