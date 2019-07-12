package main

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func main() {
	// Load vars
	dbAdminUser := os.Args[1]
	dbAdminPassword := os.Args[2]
	dbHost := os.Args[3]
	dbName := os.Args[4]
	dbTableName := os.Args[5]
	dbServiceUser := os.Args[6]
	dbServicePassword := os.Args[7]

	// Login to the DB
	connectionStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/?timeout=10s", dbAdminUser, dbAdminPassword, dbHost, "3306")

	// In this function, I know it runs only once so I can put the connection pool in the lambda
	db, connectionErr := sqlx.Connect("mysql", connectionStr)
	if connectionErr != nil {
		panic(connectionErr)
	}
	defer db.Close()

	// Create the database
	db.MustExec(fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS %s`, dbName))
	db.MustExec(fmt.Sprintf(`USE %s`, dbName))

	db.MustExec(fmt.Sprintf("CREATE USER '%s'@'%%' IDENTIFIED BY '%s';", dbServiceUser, dbServicePassword))
	db.MustExec(fmt.Sprintf("GRANT ALL ON %s.* TO '%s'@'%%';", dbName, dbServiceUser))
	// db.MustExec("FLUSH PRIVILEGES;")

	db.MustExec(fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s(
		id INT NOT NULL AUTO_INCREMENT,
		email varchar(255) DEFAULT NULL,
		date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		CONSTRAINT UNIQUE(email)
	);`, dbTableName))

	db.MustExec(fmt.Sprintf(`INSERT INTO %s (email) VALUES ("bob@example.com");`, dbTableName))
	db.MustExec(fmt.Sprintf(`INSERT INTO %s (email) VALUES ("jane@example.com");`, dbTableName))

	fmt.Println("Successfully create some users and such")
}
