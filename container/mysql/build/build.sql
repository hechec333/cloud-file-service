CREATE DATABASE IF NOT EXISTS bigdata;
USE bigdata;
CREATE TABLE `Cluster`  (
  `ClusterId` int NOT NULL,
  `UserId` int UNSIGNED NULL,
  `StackVersion` varchar(10) NOT NULL,
  `ClusterName` varchar(255) NULL,
  `CreateTime` datetime NULL,
  `UpdateTime` datetime NULL,
  PRIMARY KEY (`ClusterId`)
);

CREATE TABLE `Compotent`  (
  `CompotentId` int NOT NULL AUTO_INCREMENT,
  `StackVersion` varchar(10) NOT NULL,
  `HostId` int NOT NULL,
  `CreateTime` datetime NULL,
  `UpdateTime` datetime NULL,
  PRIMARY KEY (`CompotentId`)
);

CREATE TABLE `Host`  (
  `HostId` int NOT NULL AUTO_INCREMENT,
  `ClusterId` int NULL,
  `HostName` varchar(255) NULL,
  `RoleName` varchar(255) NULL COMMENT '是否是主节点',
  `Status` varchar(255) NULL COMMENT '当前状态',
  `SecretId` int NULL,
  `CreateTime` datetime NOT NULL,
  `UpdateTime` datetime NOT NULL ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`HostId`)
);

CREATE TABLE `Secret`  (
  `SecretId` int NOT NULL,
  `SshPrivateKey` varchar(4000) NULL,
  PRIMARY KEY (`SecretId`)
);

CREATE TABLE `Stack`  (
  `StackVersion` varchar(10) NOT NULL,
  `CompotentName` varchar(255) NOT NULL,
  `CompotentVersion` varchar(255) NULL,
  `Deployment` varchar(100) NULL,
  PRIMARY KEY (`StackVersion`, `CompotentName`)
);


CREATE TABLE `User`  (
  `UserId` int UNSIGNED NOT NULL,
  `UserAccount` varchar(255) NULL,
  `UserPassword` varchar(255) NULL,
  `UserClusterId` int UNSIGNED NULL,
  PRIMARY KEY (`UserId`)
);

ALTER TABLE `Cluster` ADD CONSTRAINT `ClusterStackVersion` FOREIGN KEY (`StackVersion`) REFERENCES `Stack` (`StackVersion`);
ALTER TABLE `Cluster` ADD FOREIGN KEY (`UserId`) REFERENCES `User` (`UserId`);
ALTER TABLE `Compotent` ADD CONSTRAINT `StackVersion` FOREIGN KEY (`StackVersion`) REFERENCES `Stack` (`StackVersion`);
ALTER TABLE `Compotent` ADD CONSTRAINT `HostId` FOREIGN KEY (`HostId`) REFERENCES `Host` (`HostId`);
ALTER TABLE `Host` ADD CONSTRAINT `ClustersId` FOREIGN KEY (`ClusterId`) REFERENCES `Cluster` (`ClusterId`);

