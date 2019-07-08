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
var dbHost = os.Getenv("DB_HOST")
var dbPort = os.Getenv("DB_PORT")
var dbAdminUser = os.Getenv("DB_ADMIN_USER")
var dbPassword = os.Getenv("DB_PASSWORD")

// Load the Env variables
var db *sqlx.DB

var userTable = `
CREATE TABLE IF NOT EXISTS user(
	id INT NOT NULL AUTO_INCREMENT,
	email varchar(255) DEFAULT NULL,
	phone varchar(10) DEFAULT NULL,
	PRIMARY KEY (id),
	CONSTRAINT UNIQUE(email),
	CONSTRAINT UNIQUE(phone)
);
`

// handleRequest Sends out the MMS Message using the Twilio Service
func handleCrudRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Actual function start and create the Database")
	// db.MustExec(`CREATE DATABASE IF NOT EXISTS ?`, dbName)
	db.MustExec(fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS %s`, dbName))
	// db.MustExec(fmt.Sprintf(`USE %s`, dbName))

	// db.MustExec(fmt.Sprintf("CREATE USER '%s' IDENTIFIED WITH AWSAuthenticationPlugin AS 'RDS';", dbServiceUser))
	// db.MustExec(fmt.Sprintf("GRANT ALL ON PRIVILEDGES ON %s.* TO '%s'@'%';", dbName, dbServiceUser))
	// db.MustExec("FLUSH PRIVILEDGES;")

	// db.MustExec(userTable)

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Hello there",
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "*",
			"Access-Control-Allow-Headers": "*",
		},
	}, nil
}

func main() {
	lambda.Start(handleCrudRequest)
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
