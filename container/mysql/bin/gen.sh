#!/bin/bash
PWD=../etc
DataPath=../data
Backup=./backup
MasterDir=../etc/master
if [[ $# < 1 ]];then
    echo "enter slave numbers @eg gen.sh @number"
    exit 
fi
if [[ ! -d $Backup ]];then 
    mkdir -p $Backup
fi
if [[ -d ../tmp ]];then 
    rm -rf ../tmp
fi
rm -rf ../etc/*
mkdir -p ../tmp
str=`date "+%Y-%m-%d %H:%M:%S"`
mkdir -p ${Backup}/$str 
mv ${DataPath}/* ${Backup}/$str
mkdir -p $MasterDir

for node in $@;do
    mkdir -p $PWD/$node
done
echo '[mysqld]
## 设置server_id,注意要唯一,注意并且不能为0,为0不可以复制日志
server-id=1
## 开启binlog
log-bin=mysql-bin
## binlog缓存
binlog_cache_size=1M
## binlog格式(mixed、statement、row,默认格式是statement)
binlog_format=mixed
##设置字符编码为utf8mb4
character-set-server = utf8mb4
collation-server = utf8mb4_unicode_ci
init_connect='SET NAMES utf8mb4'
[client]
default-character-set = utf8mb4
[mysql]
default-character-set = utf8mb4
' > $MasterDir/mysql.cnf
i=0
for node in `seq 1 $#`;do
    echo '[mysqld]
## 设置server_id,注意要唯一
server-id='$(( $node+1 ))'
## 开启binlog
log-bin=mysql-bin
## binlog缓存
binlog_cache_size=1M
## binlog格式(mixed、statement、row,默认格式是statement)
binlog_format=mixed
##设置字符编码为utf8mb4
character-set-server = utf8mb4
collation-server = utf8mb4_unicode_ci
init_connect='SET NAMES utf8mb4'
[client]
default-character-set = utf8mb4
[mysql]
default-character-set = utf8mb4
' > $PWD/${!node}/mysql.cnf
done