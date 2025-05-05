package graphql

import (
	"errors"
	m "fastcup/_pkg/models"
	q "fastcup/_pkg/queries"
)

func InitialMatchData(matchID int) (m.Match, error) {
	query := q.FullMatchQuery

	variables := map[string]int{
		"matchId": matchID,
		"gameId":  3,
	}
	var responseBody m.GetMatchStatsResponse

	if !SendGraphQLRequest(query, variables, &responseBody) {
		return m.Match{}, errors.New("match didn't initialized")
	}

	return responseBody.Data.Match, nil
}
