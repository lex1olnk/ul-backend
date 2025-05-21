package repository

import (
	"context"
	m "fastcup/_pkg/models"
	q "fastcup/_pkg/queries"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetAggregatedPlayerStats(ctx context.Context, pool *pgxpool.Pool) ([]gin.H, error) {
	query := q.GetPlayerStats

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []gin.H
	for rows.Next() {
		var s m.MapStats
		var ul *string
		clutches := [5]int{}
		if err := rows.Scan(
			&s.ID,
			&s.Nickname,
			&s.ULRating,
			&s.IMG,
			&ul,
			&s.Matches,
			&s.Kills,
			&s.Deaths,
			&s.Assists,
			&s.Headshots,
			&s.KASTScore,
			&s.FirstKill,
			&s.FirstDeath,
			&clutches[0],
			&clutches[1],
			&clutches[2],
			&clutches[3],
			&clutches[4],
			&s.Damage,
			&s.Rating,
			&s.Impact,
			&s.Rounds,
		); err != nil {
			return nil, err
		}

		clutchExp := 0
		for num := range clutches {
			clutchExp += clutches[num] * (num + 1)
		}
		if ul != nil {
			s.UL_ID = *ul
		}
		stats = append(stats, gin.H{
			"playerID":    s.ID,
			"nickname":    s.Nickname,
			"uLRating":    s.ULRating,
			"img":         s.IMG,
			"ul_id":       s.UL_ID,
			"matches":     s.Matches,
			"kills":       s.Kills,
			"deaths":      s.Deaths,
			"assists":     s.Assists,
			"headshots":   s.Headshots,
			"kast":        s.KASTScore,
			"firstKills":  s.FirstKill,
			"firstDeaths": s.FirstDeath,
			"clutchExp":   clutchExp,
			"damage":      s.Damage,
			"rating":      s.Rating,
			"impact":      s.Impact,
			"rounds":      s.Rounds,
		})
	}

	return stats, nil
}

func InitialMatchData(ctx context.Context, pool *pgxpool.Pool, playerID int, ul_id string, page int) ([]gin.H, error) {
	query := q.PlayerMatchesByUlIdQuery

	var ulIDArg interface{}
	if ul_id == "" {
		ulIDArg = nil
	} else {
		ulIDArg = ul_id
	}

	rows, err := pool.Query(ctx, query, playerID, ulIDArg, page)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var stats []gin.H
	for rows.Next() {
		var result m.APIStats
		err := rows.Scan(
			&result.MatchID,
			&result.Kills,
			&result.Deaths,
			&result.Assists,
			&result.FinishedAt,
			&result.Rating,
			&result.IsWinner,
			&result.Map,
		)

		if err != nil {
			return nil, err
		}
		stats = append(stats, gin.H{
			"matchId":    result.MatchID,
			"kills":      result.Kills,
			"deaths":     result.Deaths,
			"assists":    result.Assists,
			"finishedAt": result.FinishedAt,
			"rating":     result.Rating,
			"isWinner":   result.IsWinner,
			"map":        result.Map,
		})
	}

	return stats, nil
}

func GetAverageStats(ctx context.Context, tx pgx.Tx, playerID int) (gin.H, error) {
	query := q.GetAverageStatsQuery

	var result m.PlayerComparison
	var faceitlink *string
	err := tx.QueryRow(ctx, query, playerID).Scan(
		&result.PlayerID,
		&result.Nickname,
		&faceitlink,
		&result.ULRating,
		&result.IMG,
		&result.Kills,
		&result.Deaths,
		&result.Assists,
		&result.FirstKills,
		&result.FirstDeaths,
		&result.Flashes,
		&result.Nades,
		&result.Impact,
		&result.KAST,
		&result.Maps,
		&result.Exchanged,
		&result.V1,
		&result.V2,
		&result.V3,
		&result.V4,
		&result.V5,
		&result.Winrate,
		&result.KDPercentile,
		&result.HSPercentile,
		&result.AvgPercentile,
		&result.RatingPercentile,
		&result.TargetKD,
		&result.TargetHSRatio,
		&result.TargetAvg,
		&result.Rating,
	)

	if err != nil {
		return nil, err
	}

	if faceitlink != nil {
		parts := strings.Split(*faceitlink, "/")
		if len(parts) > 1 {
			result.Faceit = parts[len(parts)-1]
		}
	}

	return gin.H{
		"playerID":      result.PlayerID,
		"nickname":      result.Nickname,
		"faceit":        result.Faceit,
		"uLRating":      result.ULRating,
		"img":           result.IMG,
		"kills":         result.Kills,
		"deaths":        result.Deaths,
		"assists":       result.Assists,
		"firstKills":    result.FirstKills,
		"firstDeaths":   result.FirstDeaths,
		"flashes":       result.Flashes,
		"nades":         result.Nades,
		"impact":        result.Impact,
		"kast":          result.KAST,
		"maps":          result.Maps,
		"exchanged":     result.Exchanged,
		"winrate":       result.Winrate,
		"kdAdv":         result.KDPercentile,
		"hsAdv":         result.HSPercentile,
		"avgAdv":        result.AvgPercentile,
		"ratingAdv":     result.RatingPercentile,
		"TargetKD":      result.TargetKD,
		"TargetHSRatio": result.TargetHSRatio,
		"TargetAvg":     result.TargetAvg,
		"rating":        result.Rating,
		"clutches": gin.H{
			"1v1": result.V1,
			"1v2": result.V2,
			"1v3": result.V3,
			"1v4": result.V4,
			"1v5": result.V5,
		},
	}, nil
}

func GetAverageMapsStats(ctx context.Context, tx pgx.Tx, playerID int) ([]interface{}, error) {
	query := q.GetAverageMapStatsQuery
	rows, err := tx.Query(ctx, query, playerID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []interface{}
	for rows.Next() {
		var result m.MapsStatistic

		err := rows.Scan(
			&result.Map,
			&result.Matches,
			&result.Wins,
			&result.AvgRating,
			&result.Winrate,
		)

		if err != nil {
			return nil, err
		}
		stats = append(stats, gin.H{
			"map":        result.Map,
			"matches":    result.Matches,
			"wins":       result.Wins,
			"avg_rating": result.AvgRating,
			"winrate":    result.Winrate,
		})
	}

	return stats, nil
}

func GetPlayerUlTournaments(ctx context.Context, tx pgx.Tx, playerID int) ([]interface{}, error) {
	query := q.PlayerUlTournamentsQuery
	rows, err := tx.Query(ctx, query, playerID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []interface{}
	for rows.Next() {
		var id, name *string

		err := rows.Scan(
			&id,
			&name,
		)
		if id == nil {
			continue
		}
		if err != nil {
			return nil, err
		}
		stats = append(stats, gin.H{
			"id":   id,
			"name": name,
		})
	}

	return stats, nil
}
