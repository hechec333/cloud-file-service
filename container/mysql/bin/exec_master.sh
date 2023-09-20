#!/bin/bash

cp -r ../build/*.sql ../tmp
sed -i "1 i\SET @mysqlcu='$1';" ../tmp/grant.sql
sed -i "1 i\SET @mysqlcuPwd='$2';" ../tmp/grant.sql
sed -i "1 i\SET @mysqlcu='$1';" ../tmp/slave.sql
sed -i "1 i\SET @mysqlcuPwd='$2';" ../tmp/slave.sql
docker cp ../tmp/grant.sql mysql_master_1:/tmp
docker exec mysql_master_1 sh -c "mysql -u root -p123456 < /tmp/grant.sql"