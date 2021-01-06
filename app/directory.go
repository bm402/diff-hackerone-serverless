package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// GqlDirectoryRequestBody is the body for the GraphQL POST request
type GqlDirectoryRequestBody struct {
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
	Query         string                 `json:"query"`
}

// Asset is an in scope asset for a program
type Asset struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Bounty   bool   `json:"bounty"`
}

func getDirectoryFromHackerOne() map[string][]Asset {
	logger("Retrieving directory from HackerOne")

	directory := make(map[string][]Asset)
	hasNextPage := true
	cursor := ""

	for hasNextPage {
		directoryPageJSON := getDirectoryPageJSON(cursor)
		var directoryPageMap = make(map[string]interface{})
		json.Unmarshal([]byte(directoryPageJSON), &directoryPageMap)

		hasNextPage = directoryPageMap["data"].(map[string]interface{})["teams"].(map[string]interface{})["pageInfo"].(map[string]interface{})["hasNextPage"].(bool)
		cursor = directoryPageMap["data"].(map[string]interface{})["teams"].(map[string]interface{})["pageInfo"].(map[string]interface{})["endCursor"].(string)

		programs := directoryPageMap["data"].(map[string]interface{})["teams"].(map[string]interface{})["edges"].([]interface{})
		for _, program := range programs {
			programNode := program.(map[string]interface{})["node"]
			programName := programNode.(map[string]interface{})["handle"].(string)

			for _, asset := range programNode.(map[string]interface{})["in_scope_assets"].(map[string]interface{})["edges"].([]interface{}) {
				assetNode := asset.(map[string]interface{})["node"]

				if assetNode.(map[string]interface{})["eligible_for_bounty"] == nil {
					assetNode.(map[string]interface{})["eligible_for_bounty"] = false
				}

				asset := Asset{
					Name:     assetNode.(map[string]interface{})["asset_identifier"].(string),
					Type:     assetNode.(map[string]interface{})["asset_type"].(string),
					Severity: assetNode.(map[string]interface{})["max_severity"].(string),
					Bounty:   assetNode.(map[string]interface{})["eligible_for_bounty"].(bool),
				}

				directory[programName] = append(directory[programName], asset)
			}
		}
	}

	return directory
}

func getDirectoryPageJSON(cursor string) string {
	variablesJSON := `{"where":{"_and":[{"_or":[{"submission_state":{"_eq":"open"}},{"submission_state":{"_eq":"api_only"}},{"external_program":{}}]},{"_not":{"external_program":{}}},{"_or":[{"_and":[{"state":{"_neq":"sandboxed"}},{"state":{"_neq":"soft_launched"}}]},{"external_program":{}}]}]},"secureOrderBy":{"started_accepting_at":{"_direction":"DESC"}}}`
	var variables map[string]interface{}
	json.Unmarshal([]byte(variablesJSON), &variables)
	variables["cursor"] = cursor

	operationName := "DirectoryQuery"
	query := "query DirectoryQuery($cursor: String, $secureOrderBy: FiltersTeamFilterOrder, $where: FiltersTeamFilterInput) { teams(first: 100, after: $cursor, secure_order_by: $secureOrderBy, where: $where) {pageInfo { endCursor hasNextPage } edges { node { handle in_scope_assets: structured_scopes(first: 500, archived: false, eligible_for_submission: true) { edges { node { asset_type asset_identifier max_severity eligible_for_bounty }}}}}}}"

	requestBody := GqlDirectoryRequestBody{
		OperationName: operationName,
		Variables:     variables,
		Query:         query,
	}

	url := "https://hackerone.com/graphql"

	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(requestBody)

	response, err := http.Post(url, "application/json", buf)
	if err != nil {
		logger(err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logger(err)
	}

	return string(body)
}
