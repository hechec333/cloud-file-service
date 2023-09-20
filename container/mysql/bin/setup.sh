#!/bin/bash
function utime(){
    echo -n "[`date "+%Y-%m-%d %H:%M:%S"`]"
}


utime && echo " generating configure file and docker-compose file"
go run main.go compose
sleep 2s

echo " running mysql-cu cluster"
cd .. && docker-compose up -d
cd bin && echo " waitting mysql cluster configure...."
sleep 3s
echo " SettingUp mysql cluster configure..."
# go run main.go exec-cu 
# sleep 1s
utime && echo " Done..."

