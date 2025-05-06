package repository

import (
	"context"
	m "fastcup/_pkg/models"
	q "fastcup/_pkg/queries"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetAggregatedPlayerStats(ctx context.Context, pool *pgxpool.Pool) ([]gin.H, error) {
	query := q.GetPlayerStats

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	var stats []gin.H
	for rows.Next() {
		var s m.MapStats
		if err := rows.Scan(
			&s.ID,
			&s.Nickname,
			&s.ULRating,
			&s.IMG,
			&s.Matches,
			&s.Kills,
			&s.Deaths,
			&s.Assists,
			&s.Headshots,
			&s.KASTScore,
			&s.Damage,
			&s.Rating,
			&s.Impact,
			&s.Rounds,
		); err != nil {
			return nil, fmt.Errorf("data scanning error: %w", err)
		}

		stats = append(stats, gin.H{
			"playerID":  s.ID,
			"nickname":  s.Nickname,
			"uLRating":  s.ULRating,
			"img":       s.IMG,
			"matches":   s.Matches,
			"kills":     s.Kills,
			"deaths":    s.Deaths,
			"assists":   s.Assists,
			"headshots": s.Headshots,
			"kast":      s.KASTScore,
			"damage":    s.Damage,
			"rating":    s.Rating,
			"impact":    s.Impact,
			"rounds":    s.Rounds,
		})
	}

	return stats, nil
}

func InitialMatchData(ctx context.Context, tx pgx.Tx, playerID int) ([]gin.H, error) {
	query := `SELECT 
	mp.match_id,
	mp.kills,
	mp.deaths,
	mp.assists,
	mp.finished_at,
	mp.rating,
	m.map_name
FROM 
	match_players mp
LEFT JOIN 
	maps m ON m.map_id = mp.map_id
WHERE 
	player_id = $1
ORDER BY
	mp.finished_at DESC
LIMIT $2
`
	rows, err := tx.Query(ctx, query, playerID, 15)

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
			"map":        result.Map,
		})
	}

	return stats, nil
}

func GetAverageStats(ctx context.Context, tx pgx.Tx, playerID int) (gin.H, error) {
	query := q.GetAverageStatsQuery

	var result m.PlayerComparison
	err := tx.QueryRow(ctx, query, playerID).Scan(
		&result.PlayerID,
		&result.Nickname,
		&result.ULRating,
		&result.IMG,
		&result.Kills,
		&result.Deaths,
		&result.Assists,
		&result.FirstKills,
		&result.FirstDeaths,
		&result.KAST,
		&result.Maps,
		&result.WinratePercentile,
		&result.KDPercentile,
		&result.HSPercentile,
		&result.AvgPercentile,
		&result.TargetWinrate,
		&result.TargetKD,
		&result.TargetHSRatio,
		&result.TargetAvg,
	)

	if err != nil {
		return nil, err
	}

	return gin.H{
		"playerID":      result.PlayerID,
		"nickname":      result.Nickname,
		"uLRating":      result.ULRating,
		"img":           result.IMG,
		"kills":         result.Kills,
		"deaths":        result.Deaths,
		"assists":       result.Assists,
		"firstKills":    result.FirstKills,
		"firstDeaths":   result.FirstDeaths,
		"kast":          result.KAST,
		"maps":          result.Maps,
		"winrateAdv":    result.WinratePercentile,
		"kdAdv":         result.KDPercentile,
		"hsAdv":         result.HSPercentile,
		"avgAdv":        result.AvgPercentile,
		"TargetWinrate": result.TargetWinrate,
		"TargetKD":      result.TargetKD,
		"TargetHSRatio": result.TargetHSRatio,
		"TargetAvg":     result.TargetAvg,
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

	if err != nil {
		return nil, err
	}

	return stats, nil
}
