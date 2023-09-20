
CREATE DATABASE IF NOT EXISTS `test`;

USE `test`;
DROP TABLE IF EXISTS `File`;
DROP TABLE IF EXISTS `FileFolder`;
DROP TABLE IF EXISTS `Store`;
DROP TABLE IF EXISTS `User`;

CREATE TABLE `File`  (
  `FileId` bigint(11) NOT NULL AUTO_INCREMENT,
  `StoreId` int(11) NULL,
  `ParentFolderId` int(11) NULL,
  `FileName` varchar(255) NULL,
  `FilePath` varchar(255) NULL,
  `FileHash` varchar(255) NULL,
  `CreateTime` datetime NULL,
  `UpdateTime` datetime NULL ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`FileId`)
);

CREATE TABLE `FileFolder`  (
  `ID` int(11) NOT NULL AUTO_INCREMENT,
  `StoreId` int(11) NULL,
  `FileFolderPath` varchar(255) NULL,
  `FileFolderName` varchar(255) NULL,
  `ParentFolderId` int(11) NULL,
  `CreateTime` datetime NULL,
  `UpdateTime` datetime NULL ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`ID`)
);

CREATE TABLE `Store`  (
  `ID` int(11) NOT NULL,
  `CurrentUse` int(11) ZEROFILL NULL DEFAULT NULL,
  `Limits` int(11) NULL DEFAULT 1024,
  `CreateTime` datetime NOT NULL,
  `UpdateTime` datetime NOT NULL ON UPDATE CURRENT_TIMESTAMP,
  `UserId` int(11) NULL,
  PRIMARY KEY (`ID`)
);

CREATE TABLE `User`  (
  `ID` int(11) NOT NULL,
  `UserName` varchar(255) NULL,
  `UserAvator` varchar(255) NULL,
  `UserAuth` varchar(255) NULL,
  PRIMARY KEY (`ID`)
);

ALTER TABLE `File` ADD FOREIGN KEY (`StoreId`) REFERENCES `Store` (`ID`) ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE `File` ADD FOREIGN KEY (`ParentFolderId`) REFERENCES `FileFolder` (`ID`) ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE `FileFolder` ADD FOREIGN KEY (`StoreId`) REFERENCES `Store` (`ID`) ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE `Store` ADD FOREIGN KEY (`UserId`) REFERENCES `User` (`ID`);

