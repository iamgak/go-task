-- db schema
CREATE DATABASe IF NOT EXIST `go_task`;
CREATE USER 'go_task'@'localhost' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON go_task.* TO 'go_task'@'localhost';

CREATE TABLE users (
	id INT PRIMARY KEY AUTO_INCREMENT,
	email VARCHAR(255) NOT NULL,
	hash_passw VARCHAR(255) NOT NULL,
	verified_at DATETIME DEFAULT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );
  
  
CREATE TABLE users_session (
	id INT PRIMARY KEY AUTO_INCREMENT,
	user_id INT UNSIGNED NOT NULL,
	login_token VARCHAR(255) NOT NULL,
	superseded TINYINT(1) DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );

CREATE TABLE tasks (
	id INT PRIMARY KEY AUTO_INCREMENT,
	user_id INT UNSIGNED NOT NULL,
	title VARCHAR(255) DEFAULT NULL,
	description TEXT NOT NULL,
	status ENUM('Pending','In Progress','Completed') NOT NULL,
	is_deleted TINYINT(1) UNSIGNED DEFAULT 0,
	due_at DATETIME NOT NULL,
	version INT UNSIGNED DEFAULT 1,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	deleted_at DATETIME DEFAULT NULL,
	updated_at DATETIME DEFAULT NULL
  );
