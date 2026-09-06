package graphql

import (
	"context"
	"fmt"

	m "fastcup/_pkg/models"
	q "fastcup/_pkg/queries"
)

func InitialMatchData(ctx context.Context, matchID int) (m.Match, error) {
	query := q.FullMatchQuery

	variables := map[string]int{
		"matchId": matchID,
		"gameId":  3,
	}
	var responseBody m.GetMatchStatsResponse

	if err := SendGraphQLRequest(ctx, query, variables, &responseBody); err != nil {
		return m.Match{}, fmt.Errorf("match %d: %w", matchID, err)
	}

	// matches_by_pk возвращает null для несуществующего матча — без этой
	// проверки пустой матч записался бы в БД как валидный
	if responseBody.Data.Match.ID == 0 {
		return m.Match{}, fmt.Errorf("match %d not found in fastcup", matchID)
	}

	return responseBody.Data.Match, nil
}
