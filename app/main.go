package main

import (
	"context"
	"errors"
	"log"
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

func main() {
	directoryName = os.Getenv("DIRECTORY_NAME")
	region := os.Getenv("AWS_REGION")
	session, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		logger(err)
	}
	dynamoClient = dynamodb.New(session)
	lambda.Start(handler)
}

func handler(ctx context.Context, response Response) (Response, error) {
	directory := getDirectory()
	if !doesDirectoryExist() {
		createNewDirectory(directory)
		return createResponse("Created"), nil
	}
	updateDirectory(directory)
	return createResponse("Updated"), nil
}

func logger(message interface{}) {
	switch message.(type) {
	case string:
		log.Print(message)
	case error:
		sendSlackErrorNotification(message.(error))
		log.Fatal(message)
	default:
		logger(errors.New("Unknown log type"))
	}
}

func createResponse(status string) Response {
	return Response{
		Status: status,
	}
}
