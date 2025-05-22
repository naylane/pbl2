#!/bin/bash

chmod +x inicia.sh

docker-compose build
docker-compose create
docker ps -a
docker-compose start servidor1 
docker-compose logs -f
#Para o veiculo:
#docker-compose start veiculo
#docker exec -it veiculo sh
#./veiculo

# ./cmd/app/exec/inicia.sh