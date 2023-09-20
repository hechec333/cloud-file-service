
SET @query = CONCAT('CHANGE MASTER TO MASTER_HOST="master", MASTER_USER="', @mysqlcu, '", MASTER_PASSWORD="', @mysqlcuPwd, '", MASTER_LOG_FILE="', @logfile, '", MASTER_LOG_POS=', @logpos);

PREPARE stmt FROM @query;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

START SLAVE;
