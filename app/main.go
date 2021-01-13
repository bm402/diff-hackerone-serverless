package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

var (
	dynamoClient  dynamodbiface.DynamoDBAPI
	directoryName string
)

// Response is a struct for sending a response status
type Response struct {
	Status string `json:"status"`
}

// Initialises the connection to DynamoDB
func init() {
	directoryName = os.Getenv("DIRECTORY_NAME")
	region := os.Getenv("AWS_REGION")
	session, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		logger(err)
	}
	dynamoClient = dynamodb.New(session)
}

// Lambda function handler
func handler() (Response, error) {
	externalDirectory := getDirectoryFromHackerOne()
	localDirectoryCount := getLocalDirectoryCount()

	if localDirectoryCount == 0 {
		populateEmptyLocalDirectory(externalDirectory)
		return Response{
			Status: "created",
		}, nil
	}

	updateLocalDirectory(externalDirectory)
	return Response{
		Status: "updated",
	}, nil
}

// Application entry point
func main() {
	lambda.Start(handler)
}
