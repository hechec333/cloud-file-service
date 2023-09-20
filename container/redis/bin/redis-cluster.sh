#!/bin/bash

# get docker container ipAddress docker inspect --format '{{ .NetworkSettings.Networks.redis_redis_network.IPAddress }}' @containerName
# map hsotname to ip-address
# `-it` 
docker exec  redis_node1 sh -c "echo yes | redis-cli --cluster create 192.62.11.2:6379 192.62.11.3:6379 192.62.11.4:6379 192.62.11.5:6379 192.62.11.6:6379 192.62.11.7:6379 --cluster-replicas 1 -a 123456"




