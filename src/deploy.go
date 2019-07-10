package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var dbName = os.Getenv("DB_NAME")
var dbTableName = os.Getenv("DB_TABLE_NAME")
var dbHost = os.Getenv("DB_HOST")
var dbPort = os.Getenv("DB_PORT")
var dbAdminUser = os.Getenv("DB_ADMIN_USER")
var dbPassword = os.Getenv("DB_PASSWORD")
var dbServiceUser = os.Getenv("DB_SERVICE_USER")

// Load the Env variables
var db *sqlx.DB

// handleRequest Sends out the MMS Message using the Twilio Service
func initializeDatabase(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Init Successful",
	}, nil
}

func main() {
	lambda.Start(initializeDatabase)
}

func init() {
	// Login to the DB
	connectionStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/?timeout=10s", dbAdminUser, dbPassword, dbHost, dbPort)

	var err error
	db, err = sqlx.Connect("mysql", connectionStr)
	if err != nil {
		log.Fatalln(err)
	}
}
