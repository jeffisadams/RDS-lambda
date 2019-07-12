package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/rds/rdsutils"
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

	db.Close()
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(out),
	}, nil
}

func main() {
	lambda.Start(handleCrudRequest)
}

func init() {
	fmt.Println("Starting the init stuff of stuff")
	// Login to the DB
	awsCred := credentials.NewEnvCredentials()

	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_SERVICE_USER")
	region := os.Getenv("AWS_DEFAULT_REGION")
	dbName := os.Getenv("DB_NAME")
	fmt.Println(dbHost, dbUser, region, dbName)

	token, tokenErr := rdsutils.BuildAuthToken(fmt.Sprintf("%s:%d", dbHost, 3306), region, dbUser, awsCred)
	if tokenErr != nil {
		log.Fatalln(tokenErr)
	}

	connectionStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?tls=true&allowCleartextPasswords=1", dbUser, token, dbHost, dbName)

	fmt.Println("Build the token stuff")
	fmt.Println(connectionStr)

	var err error
	db, err = sqlx.Connect("mysql", connectionStr)
	if err != nil {
		log.Fatalln(err)
	}
}
