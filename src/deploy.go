package main

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Load all the env variables
	var dbName = os.Getenv("DB_NAME")
	var dbTableName = os.Getenv("DB_TABLE_NAME")
	var dbHost = os.Getenv("DB_HOST")
	var dbAdminUser = os.Getenv("DB_ADMIN_USER")
	var dbPassword = os.Getenv("DB_PASSWORD")
	var dbServiceUser = os.Getenv("DB_SERVICE_USER")

	fmt.Println("What is our env")
	fmt.Println(dbName, dbTableName, dbHost, dbAdminUser, dbPassword, dbServiceUser)

	// Login to the DB
	connectionStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/?timeout=10s", dbAdminUser, dbPassword, dbHost, "3306")
	fmt.Println(connectionStr)

	// In this function, I know it runs only once so I can put the conneciton pool in the lambda
	// db, connectionErr := sqlx.Connect("mysql", connectionStr)
	// if connectionErr != nil {
	// 	fmt.Println(err)
	// 	err = connectionErr
	// 	return
	// }
	// defer func() {
	// 	db.Close()
	// 	if r := recover(); r != nil {
	// 		fmt.Println(r)
	// 		err = errors.New("I panicked and we all died")
	// 		return
	// 	}
	// }()

	// // Create the database
	// db.MustExec(fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS %s`, dbName))
	// db.MustExec(fmt.Sprintf(`USE %s`, dbName))

	// db.MustExec(fmt.Sprintf("CREATE USER '%s' IDENTIFIED WITH AWSAuthenticationPlugin AS 'RDS';", dbServiceUser))
	// db.MustExec(fmt.Sprintf("GRANT ALL ON %s.* TO '%s'@'%%';", dbName, dbServiceUser))
	// db.MustExec("FLUSH PRIVILEGES;")

	// db.MustExec(fmt.Sprintf(`
	// CREATE TABLE IF NOT EXISTS %s(
	// 	id INT NOT NULL AUTO_INCREMENT,
	// 	email varchar(255) DEFAULT NULL,
	// 	date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	// 	PRIMARY KEY (id),
	// 	CONSTRAINT UNIQUE(email)
	// );`, dbTableName))

	// db.MustExec(fmt.Sprintf(`INSERT INTO %s (email) VALUES ("bob@example.com");`, dbTableName))
	// db.MustExec(fmt.Sprintf(`INSERT INTO %s (email) VALUES ("jane@example.com");`, dbTableName))

	fmt.Println("Got to the end of the function right before the return statement")
}
