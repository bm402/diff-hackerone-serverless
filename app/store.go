package main

import (
	"errors"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// DirectoryDocument is a MongoDB document struct for a directory document
type DirectoryDocument struct {
	Name   string  `json:"name"`
	Assets []Asset `json:"assets"`
}

// Retrieves the number of programs in the local directory
func getLocalDirectoryCount() int {
	describeTableInput := &dynamodb.DescribeTableInput{
		TableName: aws.String(directoryName),
	}
	tableMetadata, err := dynamoClient.DescribeTable(describeTableInput)
	if err != nil {
		logger(err)
	}
	// Note: item count only updates every 6 hours, not useful if an immediate or accurate count is required
	itemCount := *tableMetadata.Table.ItemCount
	logger(strconv.FormatInt(itemCount, 10) + " programs found in " + directoryName)
	return int(itemCount)
}

// Fills an empty local directory with the full program listing
func populateEmptyLocalDirectory(directory map[string][]Asset) {
	logger("Creating all programs in " + directoryName)
	for name, assets := range directory {
		directoryDocument := DirectoryDocument{
			Name:   name,
			Assets: assets,
		}
		dynamoDocument, err := dynamodbattribute.MarshalMap(directoryDocument)
		if err != nil {
			logger(err)
		}
		putItemInput := &dynamodb.PutItemInput{
			Item:      dynamoDocument,
			TableName: aws.String(directoryName),
		}
		_, err = dynamoClient.PutItem(putItemInput)
		if err != nil {
			logger(err)
		}
	}
}

// Scans the local directory and the directory retrieved from HackerOne to look for new programs,
// updated programs or deleted programs. For each change, it updates the local directory and sends
// a Slack notification detailing the change
func updateLocalDirectory(directory map[string][]Asset) {
	logger("Updating programs in " + directoryName)

	// Get full existing directory
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(directoryName),
	}
	dynamoScan, err := dynamoClient.Scan(scanInput)
	if err != nil {
		logger(err)
	}
	existingDirectory := make(map[string][]Asset)
	for _, dynamoItem := range dynamoScan.Items {
		directoryDocument := DirectoryDocument{}
		err := dynamodbattribute.UnmarshalMap(dynamoItem, &directoryDocument)
		if err != nil {
			logger(err)
		}
		existingDirectory[directoryDocument.Name] = directoryDocument.Assets
	}

	// Search for changes
	for name, assets := range directory {

		// New program
		if existingDirectory[name] == nil {
			insertNewProgram(name, assets)
			continue
		}

		// Existing program
		newAssets := []string{}
		changedAssets := []string{}
		isProgramUpdated := false
		if len(assets) != len(existingDirectory[name]) {
			isProgramUpdated = true
		}

		for _, asset := range assets {
			existingAsset, err := findAsset(asset.Name, asset.Type, existingDirectory[name])

			if err != nil {
				// New asset
				if err.Error() == "Asset not found" {
					newAssets = append(newAssets, stringifyAsset(asset))
					isProgramUpdated = true
					continue
				} else {
					logger(err)
				}
			}

			// Existing asset
			if asset.Type != existingAsset.Type || asset.Severity != existingAsset.Severity || asset.Bounty != existingAsset.Bounty {
				changedAssets = append(changedAssets, stringifyAsset(existingAsset)+" -> "+stringifyAsset(asset))
				isProgramUpdated = true
			}
		}

		// Update program
		if isProgramUpdated {
			if len(newAssets) > 0 {
				logger("New asset(s) for program \"" + name + "\" found:")
				slackMessage := "New asset(s) for program \"" + name + "\" found:\n"
				isPaid := false
				for _, newAsset := range newAssets {
					logger("\t" + newAsset)
					slackMessage += "\t" + newAsset + "\n"
					if newAsset[len(newAsset)-6:len(newAsset)-2] == "paid" {
						isPaid = true
					}
				}
				sendSlackNotification("updated-bug-bounty-programs", slackMessage, isPaid)
			}

			if len(changedAssets) > 0 {
				logger("Changed asset(s) for program \"" + name + "\" found:")
				slackMessage := "Changed asset(s) for program \"" + name + "\" found:\n"
				isPaid := false
				for _, changedAsset := range changedAssets {
					logger("\t" + changedAsset)
					slackMessage += "\t" + changedAsset + "\n"
					if changedAsset[len(changedAsset)-6:len(changedAsset)-2] == "paid" {
						isPaid = true
					}
				}
				sendSlackNotification("updated-bug-bounty-programs", slackMessage, isPaid)
			}

			if len(assets)-len(newAssets) < len(existingDirectory[name]) {
				logger("Deleting dead asset(s) from program \"" + name + "\"")
			}

			updateProgram(name, assets)
		}

		// Remove existing program from list to remove
		delete(existingDirectory, name)
	}

	// Delete dead programs
	for name := range existingDirectory {
		deleteDeadProgram(name)
	}
}

// Inserts a new program to the local directory and sends a Slack notification detailing the new program
func insertNewProgram(name string, assets []Asset) {
	logger("New program \"" + name + "\" found with the following assets:")
	slackMessage := "New program \"" + name + "\" found with the following assets:\n"
	isPaid := false

	for _, asset := range assets {
		logger("\t" + stringifyAsset(asset))
		slackMessage += "\t" + stringifyAsset(asset) + "\n"
		if asset.Bounty {
			isPaid = true
		}
	}
	sendSlackNotification("new-bug-bounty-programs", slackMessage, isPaid)

	directoryDocument := DirectoryDocument{
		Name:   name,
		Assets: assets,
	}
	dynamoDocument, err := dynamodbattribute.MarshalMap(directoryDocument)
	if err != nil {
		logger(err)
	}

	putItemInput := &dynamodb.PutItemInput{
		Item:      dynamoDocument,
		TableName: aws.String(directoryName),
	}
	_, err = dynamoClient.PutItem(putItemInput)
	if err != nil {
		logger(err)
	}
}

// Updates a program in the local directory
func updateProgram(name string, assets []Asset) {
	assetsMap := map[string]interface{}{":assets": assets}
	dynamoAssets, err := dynamodbattribute.MarshalMap(assetsMap)
	if err != nil {
		logger(err)
	}

	updateItemInput := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: dynamoAssets,
		TableName:                 aws.String(directoryName),
		Key: map[string]*dynamodb.AttributeValue{
			"name": {
				S: aws.String(name),
			},
		},
		UpdateExpression: aws.String("set assets = :assets"),
	}

	_, err = dynamoClient.UpdateItem(updateItemInput)
	if err != nil {
		logger(err)
	}
}

// Deletes a program that is no longer open or public in the local directory
func deleteDeadProgram(name string) {
	logger("Deleting dead program \"" + name + "\"")

	deleteItemInput := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"name": {
				S: aws.String(name),
			},
		},
		TableName: aws.String(directoryName),
	}

	_, err := dynamoClient.DeleteItem(deleteItemInput)
	if err != nil {
		logger(err)
	}
}

// Creates a one-line string detailing an asset for a program
func stringifyAsset(asset Asset) string {
	str := "[ " + asset.Name + " | " + asset.Type + " | " + asset.Severity + " | "
	if asset.Bounty {
		str += "paid"
	} else {
		str += "free"
	}
	return str + " ]"
}

// Searches for a given asset in a list of assets
func findAsset(name string, assetType string, assets []Asset) (Asset, error) {
	for _, asset := range assets {
		if asset.Name == name && asset.Type == assetType {
			return asset, nil
		}
	}
	return Asset{}, errors.New("Asset not found")
}
