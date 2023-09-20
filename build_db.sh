#!/bin/bash
docker cp ./file.sql mysql_master_1:/tmp
docker exec mysql_master_1 sh -c "mysql -uroot -p123456 < /tmp/file.sql"