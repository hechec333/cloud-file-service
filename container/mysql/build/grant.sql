
SET @query1 = CONCAT('CREATE USER "', @mysqlcu, '"@"%" IDENTIFIED BY "', @mysqlcuPwd, '"');
PREPARE stmt FROM @query1;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @query2 = CONCAT('GRANT REPLICATION SLAVE ON *.* TO "', @mysqlcu, '"@"%"');
PREPARE stmt FROM @query2;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

FLUSH PRIVILEGES;
