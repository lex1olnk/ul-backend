package repository

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func GetUlTournaments(ctx context.Context, tx pgx.Tx) ([]gin.H, error) {
	query := `SELECT 
			id, name
		FROM 
			ul_tournaments
		ORDER BY
			name
	`

	rows, err := tx.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var stats []gin.H

	for rows.Next() {
		var result struct {
			id   string
			name string
		}
		err := rows.Scan(
			&result.id,
			&result.name,
		)

		if err != nil {
			return nil, err
		}
		stats = append(stats, gin.H{"id": result.id, "name": result.name})
	}

	return stats, nil
}

func PostUlTournaments(ctx context.Context, tx pgx.Tx, name string) (string, error) {
	query := `
        INSERT INTO ul_tournaments (name)
        VALUES ($1)
        RETURNING id;
    `

	var tournamentID string
	err := tx.QueryRow(ctx, query, name).Scan(&tournamentID)
	if err != nil {
		return "", err
	}

	return tournamentID, nil
}

func PostUlPlayerPick(ctx context.Context, tx pgx.Tx, id string, nickname string, pick int) error {
	query := `
        INSERT INTO player_tournament_picks (player_id, ul_tournament_id, pick_number)
        VALUES ($1, $2, $3)
        RETURNING id;
    `

	var tournamentID string
	err := tx.QueryRow(ctx, query, id, nickname, pick).Scan(&tournamentID)
	if err != nil {
		return err
	}

	return nil
}
