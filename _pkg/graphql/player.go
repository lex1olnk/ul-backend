package graphql

import (
	"errors"
	m "fastcup/_pkg/models"
	q "fastcup/_pkg/queries"
)

func GetMatchClutches(matchID int, st *m.MatchApi) error {
	query := q.MatchCluchesQuery

	variables := map[string]int{
		"matchId": matchID,
	}

	var responseBody m.GraphQLClutchResponse
	if !SendGraphQLRequest(query, variables, &responseBody) {
		return errors.New("failed to get clutches from fastcup")
	}
	st.InitClutch(responseBody.Data.Clutches)
	return nil
}

func GetMatchKills(matchID int, stats *m.MatchApi) error {
	query := q.MapMatchKillsQuery

	variables := map[string]int{
		"matchId": matchID,
		//"userId":  0, // Замените на нужный userId, если требуется
	}

	var responseBody m.GraphQLKillsResponse
	if !SendGraphQLRequest(query, variables, &responseBody) {
		return errors.New("failed to get kills from fastcup")
	}
	stats.InitKills(responseBody.Data.Kills)
	return nil
}

func GetMatchDamages(matchID int, stats *m.MatchApi) error {
	query := q.MatchDamageQuery

	variables := map[string]int{
		"matchId": matchID,
		//"userId":  0, // Замените на нужный userId, если требуется
	}

	// Декодируем JSON-ответ
	var responseBody m.GraphQLDamagesResponse
	if !SendGraphQLRequest(query, variables, &responseBody) {
		return errors.New("failed to get damages from fastcup")
	}

	stats.InitDamage(responseBody.Data.Damages)
	return nil
}
