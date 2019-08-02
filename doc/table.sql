CREATE TABLE `tbl_file` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `file_sha1` char(40) NOT NULL DEFAULT '' COMMENT 'filehash',
  `file_name` varchar(256) NOT NULL DEFAULT '' COMMENT 'file name',
  `file_size` bigint(20) DEFAULT '0' COMMENT 'file size',
  `file_addr` varchar(1024) NOT NULL DEFAULT '' COMMENT 'file location',
  `create_at` datetime default NOW() COMMENT 'created data',
  `update_at` datetime default NOW() on update current_timestamp() COMMENT 'update date',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT 'OK/Declined/Deleted)',
  `ext1` int(11) DEFAULT '0' COMMENT 'backup field1',
  `ext2` text COMMENT 'backup field2',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_file_hash` (`file_sha1`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `tbl_user` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_name` varchar(64) NOT NULL DEFAULT '' ,
  `user_pwd` varchar(256) NOT NULL DEFAULT '' ,
  `email` varchar(64) DEFAULT '' ,
  `phone` varchar(128) DEFAULT '' ,
  `email_validated` tinyint(1) DEFAULT 0 ,
  `phone_validated` tinyint(1) DEFAULT 0 ,
  `signup_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `last_active` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `profile` text COMMENT 'user properties',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '账户状态(启用/禁用/锁定/标记删除等)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_phone` (`phone`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4;