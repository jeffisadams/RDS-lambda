package main

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func main() {
	// Load all the cli variables
	var dbAdminUser = os.Args[1]
	var dbPassword = os.Args[2]
	var dbHost = os.Args[3]
	var dbName = os.Args[4]
	var dbTableName = os.Args[5]
	var dbServiceUser = os.Args[6]

	// Login to the DB
	connectionStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/?timeout=10s", dbAdminUser, dbPassword, dbHost, "3306")
	fmt.Println(connectionStr)

	// In this function, I know it runs only once so I can put the conneciton pool in the lambda
	db, err := sqlx.Connect("mysql", connectionStr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	// Create the database
	db.MustExec(fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS %s`, dbName))
	db.MustExec(fmt.Sprintf(`USE %s`, dbName))

	db.MustExec(fmt.Sprintf("CREATE USER '%s' IDENTIFIED WITH AWSAuthenticationPlugin AS 'RDS';", dbServiceUser))
	db.MustExec(fmt.Sprintf("GRANT ALL ON %s.* TO '%s'@'%%';", dbName, dbServiceUser))
	db.MustExec("FLUSH PRIVILEGES;")

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

	fmt.Println("Got to the end of the function right before the return statement")
}
