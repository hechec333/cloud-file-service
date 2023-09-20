#!/bin/bash

target=../../var/.redis
if [[  -f $target ]];then
    rm -f $target 
    touch $target
fi 
for port in $@;do 
    echo -n "${host}:${port} " >> $target 
    redis-cli -c -h ${host} -p ${port} -a ${password} info | grep role | awk -F : '{print $2}' >> $target
done 


