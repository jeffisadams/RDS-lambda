package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// UserTransaction Creates a struct for the return values from out SQL table
type UserTransaction struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Date  string `json:"date"`
}

// Load the Env variables
var db *sqlx.DB

// handleRequest Sends out the MMS Message using the Twilio Service
func handleCrudRequest() (events.APIGatewayProxyResponse, error) {
	fmt.Println("Actual function start and create the Database")
	dbTableName := os.Getenv("DB_TABLE_NAME")
	transactions := []UserTransaction{}
	err := db.Select(&transactions, fmt.Sprintf(`SELECT * from %s`, dbTableName))
	if err != nil {
		fmt.Println(err)
	}

	out, serializationErr := json.Marshal(transactions)
	if serializationErr != nil {
		fmt.Println(serializationErr)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(out),
	}, nil
}

func main() {
	lambda.Start(handleCrudRequest)
}

func init() {
	// Login to the DB
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_SERVICE_USER")
	dbPassword := os.Getenv("DB_SERVICE_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connectionStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?timeout=10s", dbUser, dbPassword, dbHost, "3306", dbName)
	var err error
	db, err = sqlx.Connect("mysql", connectionStr)
	if err != nil {
		log.Fatalln(err)
	}
}
