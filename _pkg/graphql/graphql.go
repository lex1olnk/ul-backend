package graphql

import (
	"bytes"
	"encoding/json"
	m "fastcup/_pkg/models"
	"net/http"
)

func SendGraphQLRequest(query string, variables map[string]int, responseBody interface{}) bool {
	requestBody := m.GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	// Кодируем тело запроса в JSON
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return false
	}

	resp, err := http.Post("https://hasura.fastcup.net/v1/graphql", "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return false
	}

	return true
}
