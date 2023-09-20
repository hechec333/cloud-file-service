#!/bin/bash


function utime(){
    echo -n "[`date "+%Y-%m-%d %H:%M:%S"`]"
}

utime
echo " configure service.ini file ,wait for done..."
go run main.go gen || exit 1


utime 
echo " --> done..." && sleep 2s 
cd .. && docker-compose up -d 

cd bin && echo ">>> SettingUp Redis-clusters ... "

go run main.go link || exit 1

sleep 1s
go run main.go var && echo ">>> Success... " 


