CREATE TABLE `user` (
  `user_name` varchar(255) NOT NULL,
  `password` char(32) NOT NULL,
  PRIMARY KEY (`user_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `user_profile` (
  `user_name` varchar(255) NOT NULL,
  `nick_name` varchar(255) DEFAULT NULL,
  `pic_name` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`user_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
