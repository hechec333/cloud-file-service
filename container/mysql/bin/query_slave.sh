#!/bin/bash
docker cp ../build/status.sql mysql_master_1:/tmp
#docker exec -it mysql_master_1 bin/bash -c "mysql -uroot -p123456 < /tmp/status.sql" | sed -n "3,4p" | awk '{print $2}'
rm -f .tmp
docker exec mysql_master_1 bin/bash -c "mysql -uroot -p123456 < /tmp/status.sql" | sed -n "2,3p"| awk '{print $2}' > .tmp
logfile=`sed -n "1p" .tmp`
logpos=`sed -n "2p" .tmp`
sed -i "1 i\SET @logfile='$logfile';" ../tmp/slave.sql
sed -i "1 i\SET @logpos='$logpos';" ../tmp/slave.sql
for node in $@;do
    docker cp ../tmp/slave.sql mysql_${node}_1:/tmp
# pay the fucking attention about the arg "-it",if you don't fuck use bash,fucking don't use `-it` 
    docker exec mysql_${node}_1 sh -c "mysql -uroot -p123456 < /tmp/slave.sql" 
done
