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

// var dbReadOnlyHost = os.Getenv("DB_HOST")
// var dbPort = os.Getenv("DB_PORT")
// var dbServiceUser = os.Getenv("DB_SERVICE_USER")
// var awsRegion = os.Getenv("AWS_REGION")
var dbName = os.Getenv("DB_NAME")
var dbTableName = os.Getenv("DB_TABLE_NAME")

// Load the Env variables
var db *sqlx.DB

// handleRequest Sends out the MMS Message using the Twilio Service
func handleCrudRequest() (events.APIGatewayProxyResponse, error) {
	fmt.Println("Actual function start and create the Database")
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
	awsCred := credentials.NewEnvCredentials()
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_SERVICE_USER")
	region := os.Getenv("AWS_DEFAULT_REGION")
	token, err := rdsutils.BuildAuthToken(fmt.Sprintf("%s:%d", dbHost, 3306), region, dbUser, awsCred)
	connectionStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?tls=true&allowCleartextPasswords=1", dbUser, token, dbHost, dbName)

	db, err = sqlx.Connect("mysql", connectionStr)
	if err != nil {
		log.Fatalln(err)
	}
}
