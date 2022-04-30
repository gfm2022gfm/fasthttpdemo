#!/bin/bash

docker-compose down
docker-compose up --build -d


docker rmi -f $(docker images | grep "none" | awk '{print $3}')
