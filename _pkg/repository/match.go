package repository

import (
	"context"
	"fastcup/_pkg/graphql"
	m "fastcup/_pkg/models"
	s "fastcup/_pkg/service"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateMatch(pool *pgxpool.Pool, matchID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Начало транзакции
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // Безопасный откат при ошибках

	// 1. Проверка существования матча
	fmt.Println("1. Проверка существования матча")
	exists, err := checkMatchExists(ctx, tx, matchID)
	if err != nil {
		return err
	}
	if exists {
		return nil // или возвращать ошибку, если матч уже существует
	}

	fmt.Println("2. Получение данных матча")
	// 2. Получение данных матча
	match, err := getMatchData(matchID)

	if err != nil {
		return fmt.Errorf("failed to get match data: %w", err)
	}
	stats := m.MatchApi{}
	stats.AddMatchData(match)
	stats.InitPlayers(match.Members)

	/*
		fmt.Println("3. Сохранение команд")
		// 3. Сохранение команд
		if err := saveTeams(ctx, tx, match.Teams); err != nil {
			return err
		}
	*/
	fmt.Println("4. Сохранение игроков")

	// 4. Сохранение игроков

	err = savePlayers(ctx, tx, stats.Players)

	if err != nil {
		return err
	}

	fmt.Println("5. Сохранение основной информации матча")
	// 5. Сохранение основной информации матча
	if err := saveMatchInfo(ctx, tx, matchID, match); err != nil {
		return err
	}
	fmt.Println("5.2. Сохранение карт")
	if err := saveMatchMapsInfo(ctx, tx, stats); err != nil {
		return err
	}

	// 6. Получение дополнительной статистики
	fmt.Println("6. Получение дополнительной статистики")
	if err := getMatchStatistics(matchID, &stats); err != nil {
		return err
	}
	fmt.Println("7. Обработка дополнительной статистики")
	// 7. Обработка дополнительной статистики

	s.ProcessStatistic(&stats)
	fmt.Println("8. Сохранение статистики игроков")
	// 8. Сохранение статистики игроков
	if err := savePlayersStats(ctx, tx, matchID, stats.Maps); err != nil {
		return err
	}

	// Фиксация транзакции
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	return nil
}

// Вспомогательные функции

func checkMatchExists(ctx context.Context, tx pgx.Tx, matchID int) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM matches WHERE match_id = $1)",
		matchID,
	).Scan(&exists)
	return exists, err
}

func getMatchData(matchID int) (m.Match, error) {
	match, err := graphql.InitialMatchData(matchID)

	return match, err
}

/*
	func saveTeams(ctx context.Context, tx pgx.Tx, teams []m.MatchTeam) error {
		for _, team := range teams {
			_, err := tx.Exec(ctx,
				`INSERT INTO teams (team_id, name)
						VALUES ($1, $2)
						ON CONFLICT (team_id) DO NOTHING`,
				team.ID,
				team.Name, // Используем настоящее имя команды
			)
			if err != nil {
				return fmt.Errorf("failed to insert team %d: %w", team.ID, err)
			}
		}
		return nil
	}
*/
func savePlayers(ctx context.Context, tx pgx.Tx, players []m.PlayerInit) error {
	for _, player := range players {
		_, err := tx.Exec(ctx,
			`INSERT INTO players (player_id, ul_rating, nickname, img)
					VALUES ($1, $2, $3, $4)
					ON CONFLICT (player_id) DO NOTHING`,
			player.ID,
			0.0,
			player.Nickname,
			player.IMG,
		)
		if err != nil {
			return fmt.Errorf("failed to upsert player %d: %w", player.ID, err)
		}
	}
	return nil
}

func saveMatchInfo(ctx context.Context, tx pgx.Tx, matchID int, match m.Match) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO matches 
			(match_id, best_of)
			VALUES ($1, $2)`,
		matchID,
		match.BestOf,
	)
	return err
}

func saveMatchMapsInfo(ctx context.Context, tx pgx.Tx, st m.MatchApi) error {
	err := error(nil)
	for _, m := range st.Maps {
		_, err = tx.Exec(ctx,
			`INSERT INTO maps
				(map_id, map_name)
				VALUES ($1, $2)
				ON CONFLICT (map_id) DO NOTHING`,
			m.MapID,
			m.MapName,
		)
	}
	return err
}

func getMatchStatistics(matchID int, stats *m.MatchApi) error {
	if err := graphql.GetMatchKills(matchID, stats); err != nil {
		return err
	}
	if err := graphql.GetMatchDamages(matchID, stats); err != nil {
		return err
	}
	if err := graphql.GetMatchClutches(matchID, stats); err != nil {
		return err
	}
	return nil
}

func savePlayersStats(ctx context.Context, tx pgx.Tx, matchID int, maps map[int]m.Stats) error {
	for _, m := range maps {
		for _, player := range m.MapStats {
			player.Rounds = m.Rounds
			player.CalculateDerivedStats()
			_, err := tx.Exec(ctx,
				`INSERT INTO match_players 
            (match_id, player_id, map_id, is_winner, kills, deaths, assists, headshots, exchanged, 
            firstdeaths, firstkills, damage, multi_kills, clutches, impact, rating, kastscore, flash_assists, grenades_damages, rounds, started_at, finished_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)`,
				matchID,
				player.ID,
				m.MapID,
				player.IsWinner,
				player.Kills,
				player.Deaths,
				player.Assists,
				player.Headshots,
				player.Exchanged,
				player.FirstDeath,
				player.FirstKill,
				player.Damage,
				pgtype.FlatArray[int](player.MultiKills[:]),
				pgtype.FlatArray[int](player.Clutches[:]),
				player.Impact,
				player.Rating,
				player.KASTScore,
				player.Flashes,
				player.NadesDamage,
				m.Rounds,
				player.StartedAt,
				player.FinishedAt,
			)
			if err != nil {
				return fmt.Errorf("failed to insert stats for player %d: %w", player.ID, err)
			}
		}
	}
	return nil
}

// Вспомогательные функции

func GetPlayerMatches(ctx context.Context, pool *pgxpool.Pool, playerID int, limit int) ([]m.APIStats, error) {
	query := `
		SELECT 
			mp.match_id,
			m.map,
			mp.kills,
			mp.deaths,
			mp.assists,
			m.rounds,
			mp.rating, 
			m.finished_at
		FROM 
			match_players AS mp
		JOIN 
			matches AS m ON mp.match_id = m.id
		WHERE 
			mp.player_id = $1
		ORDER BY 
			mp.created_at DESC
		LIMIT $2;
  `

	rows, err := pool.Query(ctx, query, playerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []m.APIStats
	for rows.Next() {
		var stats m.APIStats

		err := rows.Scan(
			&stats.MatchID,
			&stats.Map,
			&stats.Kills,
			&stats.Deaths,
			&stats.Assists,
			&stats.Rounds,
			&stats.Rating,
			&stats.FinishedAt,
		)
		if err != nil {
			return nil, err
		}
		matches = append(matches, stats)
	}

	return matches, nil
}
