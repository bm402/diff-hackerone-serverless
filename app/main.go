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

func handler() (Response, error) {
	directory := getDirectory()

	createResponse := func(status string) Response {
		return Response{
			Status: status,
		}
	}

	if !doesDirectoryExist() {
		createNewDirectory(directory)
		return createResponse("created"), nil
	}

	updateDirectory(directory)
	return createResponse("updated"), nil
}

func main() {
	lambda.Start(handler)
}
