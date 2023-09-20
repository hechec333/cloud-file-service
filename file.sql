GRANT ALL PRIVILEGES ON *.* TO 'mysqlcu'@'%' IDENTIFIED BY '123456' WITH GRANT OPTION;
FLUSH PRIVILEGES;
CREATE DATABASE IF NOT EXISTS `test`;

USE `test`;
DROP TABLE IF EXISTS `File`;
DROP TABLE IF EXISTS `FileFolder`;
DROP TABLE IF EXISTS `Store`;
DROP TABLE IF EXISTS `User`;
DROP TABLE IF EXISTS `Symbol`;
DROP TABLE IF EXISTS `Grant`;
DROP TABLE IF EXISTS `WhiteList`;
DROP TABLE IF EXISTS `Credentials`;

CREATE TABLE `File`  (
  `FileId` bigint(11) NOT NULL AUTO_INCREMENT,
  `StoreId` int(11) NULL,
  `ParentFolderId` int(11) NOT NULL,
  `FileName` varchar(255) NULL,
  `FilePath` varchar(255) NULL,
  `FileHash` varchar(255) NULL,
  `CreateTime` datetime NULL,
  `FileType` varchar(255) NOT NULL, --r只读,rw可读可写
  `AuthMode` int(4) NOT NULL, --可见性设置 0:私有，1.共享
  `GrantId` int(11) NULL, --为0代表默认权限，
  `FileSize` varchar(255) NULL,
  `Persiter` varchar(255) NULL, --oss类型
  `UpdateTime` datetime NULL ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`FileId`)
);

CREATE TABLE `FileFolder`  (
  `ID` int(11) NOT NULL AUTO_INCREMENT,
  `StoreId` int(11) NULL,
  `FileFolderName` varchar(255) NULL,
  `ParentFolderId` int(11) NOT NULL, 
  `Persiter` varchar(255) NULL, -- oss类型
  `CreateTime` datetime NULL,
  `UpdateTime` datetime NULL ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`ID`)
);

CREATE TABLE `Store`  (
  `ID` int(11) NOT NULL,
  `CurrentUse` int(11) ZEROFILL NULL DEFAULT NULL,
  `Limits` int(11) NULL DEFAULT 1024,
  `CreateTime` datetime NOT NULL,
  `Persiter` varchar(255) NOT NULL,
  `UpdateTime` datetime NOT NULL ON UPDATE CURRENT_TIMESTAMP,
  `UserId` int(11) NULL,
  PRIMARY KEY (`ID`)
);

CREATE TABLE `User`  (
  `ID` int(11) NOT NULL,
  `UserEmail` varchar(255) NULL,
  `UserName` varchar(255) NULL,
  `UserPassword` varchar(255) NULL,
  `UserAvator` varchar(255) NULL,
  PRIMARY KEY (`ID`)
);

CREATE TABLE `Symbol`(
  `ID` INT(11) AUTO_INCREMENT,
  `SourceId` INT(11) NOT NULL,
  `DstId` INT(11) NOT NULL,
  PRIMARY KEY(`ID`)
);


CREATE TABLE `Grant`  (
  `GrantId` int NOT NULL,
  `ObjectId` int NOT NULL,
  `ObjectType` varchar(255) NULL,
  `GrantType` varchar(255) NULL COMMENT "r读,rw读写",
  `OwnerId` int NULL COMMENT '指向userid',
  PRIMARY KEY (`GrantId`)
);

CREATE TABLE `WhiteList`  (
  `ID` int NOT NULL AUTO_INCREMENT,
  `GrantId` int NULL,
  `GuestId` int NULL, --设置0代表开放到所有用户
  PRIMARY KEY (`ID`)
);

CREATE TABLE `Credentials`  (
  `ClientId` varchar(255) NOT NULL,
  `ClientSecret` varchar(255) NULL,
  `RedirectUrl` varchar(512) NULL,
  PRIMARY KEY (`ClientId`)
);


-- 外键
ALTER TABLE `File` ADD FOREIGN KEY (`StoreId`) REFERENCES `Store` (`ID`) ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE `File` ADD FOREIGN KEY (`ParentFolderId`) REFERENCES `FileFolder` (`ID`) ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE `FileFolder` ADD FOREIGN KEY (`StoreId`) REFERENCES `Store` (`ID`) ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE `Store` ADD FOREIGN KEY (`UserId`) REFERENCES `User` (`ID`);
ALTER TABLE `WhiteList` ADD FOREIGN KEY (`GrantId`) REFERENCES `Grant` (`GrantId`) ON DELETE CASCADE;

-- 索引
CREATE UNIQUE INDEX `index_folder_parent` ON `FileFolder`(`ParentFolderId`)
CREATE INDEX `index_folder_name` ON `FileFolder`(`FileFolderName`)
CREATE UNIQUE INDEX `index_file_parent` ON `File`(`ParentFolderId`)
CREATE INDEX `index_symbol` ON `Symbol`(`SourceId`,`DstId`)