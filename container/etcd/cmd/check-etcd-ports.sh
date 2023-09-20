#!/bin/bash

# netstat -lnput | grep docker-proxy | awk ' $1 == tcp && $4 ~ /^[0-9.]+:3[0-9]{3}$/ {print $0}'
for port in {3379,3380,3381};do
netstat -lnput | grep $port > /dev/null
if [ $? -ne 0 ];then 
    echo 'etcd port:'$port' miss'
else 
    echo 'etcd port:'$port' running'
fi
done 

