SET @logpos='753';
SET @logfile='mysql-bin.000003';
SET @mysqlcuPwd='123456';
SET @mysqlcu='mysqlcu';

SET @query = CONCAT('CHANGE MASTER TO MASTER_HOST="master", MASTER_USER="', @mysqlcu, '", MASTER_PASSWORD="', @mysqlcuPwd, '", MASTER_LOG_FILE="', @logfile, '", MASTER_LOG_POS=', @logpos);

PREPARE stmt FROM @query;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

START SLAVE;
