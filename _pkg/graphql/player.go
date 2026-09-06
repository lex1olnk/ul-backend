package graphql

import (
	"context"
	"fmt"

	m "fastcup/_pkg/models"
	q "fastcup/_pkg/queries"
)

func GetMatchClutches(ctx context.Context, matchID int, st *m.MatchApi) error {
	query := q.MatchCluchesQuery

	variables := map[string]int{
		"matchId": matchID,
	}

	var responseBody m.GraphQLClutchResponse
	if err := SendGraphQLRequest(ctx, query, variables, &responseBody); err != nil {
		return fmt.Errorf("failed to get clutches for match %d: %w", matchID, err)
	}
	st.InitClutch(responseBody.Data.Clutches)
	return nil
}

func GetMatchKills(ctx context.Context, matchID int, stats *m.MatchApi) error {
	query := q.MapMatchKillsQuery

	variables := map[string]int{
		"matchId": matchID,
	}

	var responseBody m.GraphQLKillsResponse
	if err := SendGraphQLRequest(ctx, query, variables, &responseBody); err != nil {
		return fmt.Errorf("failed to get kills for match %d: %w", matchID, err)
	}
	stats.InitKills(responseBody.Data.Kills)
	return nil
}

func GetMatchDamages(ctx context.Context, matchID int, stats *m.MatchApi) error {
	query := q.MatchDamageQuery

	variables := map[string]int{
		"matchId": matchID,
	}

	var responseBody m.GraphQLDamagesResponse
	if err := SendGraphQLRequest(ctx, query, variables, &responseBody); err != nil {
		return fmt.Errorf("failed to get damages for match %d: %w", matchID, err)
	}
	stats.InitDamage(responseBody.Data.Damages)
	return nil
}
