#!/bin/bash
if [ ! -d "etcd" ];then 
    echo "Creating Volume Dir in : etcd"
    mkdir -p etcd/etcd{1..3}/data
fi 
if [[ $# -gt 1 ]];then 
    if [[ $1 = "clean" ]];then 
        for node in {1..3};do
            rm -rf 'etcd/ectd'${node}'/data/*' 
        done
        echo "Cleanning all data-file"
    else 
        echo "unsupport flag: $1"
    fi
fi 

docker-compose up -d 
