package repository

import (
	"context"
	"fastcup/_pkg/graphql"
	m "fastcup/_pkg/models"
	s "fastcup/_pkg/service"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/api/sheets/v4"
)

func CreateMatch(ctx context.Context, tx pgx.Tx, matchID int, tournamentId *string) error {
	// 1. Проверка существования матча
	//fmt.Println("1. Проверка существования матча")
	exists, err := checkMatchExists(ctx, tx, matchID)
	if err != nil {
		return err
	}
	if exists {
		return nil // или возвращать ошибку, если матч уже существует
	}

	//fmt.Println("2. Получение данных матча")
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
	//fmt.Println("4. Сохранение игроков")

	// 4. Сохранение игроков

	err = savePlayers(ctx, tx, stats.Players)

	if err != nil {
		return err
	}

	//fmt.Println("5. Сохранение основной информации матча")
	// 5. Сохранение основной информации матча
	if err := saveMatchInfo(ctx, tx, matchID, match, tournamentId); err != nil {
		return err
	}
	//fmt.Println("5.2. Сохранение карт")
	if err := saveMatchMapsInfo(ctx, tx, stats); err != nil {
		return err
	}

	// 6. Получение дополнительной статистики
	//fmt.Println("6. Получение дополнительной статистики")
	if err := getMatchStatistics(matchID, &stats); err != nil {
		return err
	}
	//fmt.Println("7. Обработка дополнительной статистики")
	// 7. Обработка дополнительной статистики

	s.ProcessStatistic(&stats)
	//fmt.Println("8. Сохранение статистики игроков")
	// 8. Сохранение статистики игроков
	if err := savePlayersStats(ctx, tx, matchID, stats.Maps); err != nil {
		return err
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

func saveMatchInfo(ctx context.Context, tx pgx.Tx, matchID int, match m.Match, tournamentId *string) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO matches 
			(match_id, best_of, ul_tournament_id)
			VALUES ($1, $2, $3)`,
		matchID,
		match.BestOf,
		tournamentId,
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

// Функция записи в Google Sheets
func WriteToGoogleSheets(ul_id string, sheetName string, players []gin.H, srv *sheets.Service, spreadsheetID string) error {
	// 1. Проверка существования листа и создание при необходимости
	spreadsheet, err := srv.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	sheetExists := false
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			sheetExists = true
			break
		}
	}

	if !sheetExists {
		// Создаем новый лист
		addSheetReq := &sheets.Request{
			AddSheet: &sheets.AddSheetRequest{
				Properties: &sheets.SheetProperties{
					Title: sheetName,
				},
			},
		}

		batchUpdateReq := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{addSheetReq},
		}

		_, err := srv.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateReq).Do()
		if err != nil {
			return fmt.Errorf("failed to create sheet: %w", err)
		}
		log.Printf("Created new sheet: %s", sheetName)
	}
	// Подготовка заголовков столбцов
	headers := []string{
		"Nickname", "Pick Number", "Matches Played", "Kills", "Deaths", "Assists", "Headshots",
		"KAST Score", "First Kills", "First Deaths", "ClutchExp", "Total Damage",
		"Impact", "Total Rounds", "Rating",
	}

	values := [][]interface{}{}
	values = append(values, convertToInterfaceSlice(headers))

	// Подготовка данных игроков
	for _, p := range players {
		row := []interface{}{
			getString(p, "nickname"),
			getInt(p, "pick_number"),
			getInt(p, "matches"),
			getInt(p, "kills"),
			getInt(p, "deaths"),
			getInt(p, "assists"),
			getInt(p, "headshots"),
			getFloat(p, "kast"),
			getInt(p, "firstKills"),
			getInt(p, "firstDeaths"),
			getInt(p, "clutchExp"),
			getFloat(p, "damage"),
			getFloat(p, "impact"),
			getInt(p, "rounds"),
			getFloat(p, "rating"),
		}
		values = append(values, row)
	}

	// Определение диапазона записи
	rangeData := fmt.Sprintf("%s!A:X", sheetName)
	valueRange := &sheets.ValueRange{
		Values: values,
	}

	// Очистка листа перед записью новых данных
	clearRange := fmt.Sprintf("%s!A:Z", sheetName)
	if _, err := srv.Spreadsheets.Values.Clear(spreadsheetID, clearRange, &sheets.ClearValuesRequest{}).Do(); err != nil {
		return fmt.Errorf("failed to clear sheet: %w", err)
	}

	// Записываем новые данные
	_, err = srv.Spreadsheets.Values.Update(
		spreadsheetID,
		rangeData,
		valueRange,
	).ValueInputOption("USER_ENTERED").Do()

	return err
}

// Вспомогательные функции для безопасного извлечения данных
func getInt(data gin.H, key string) int {
	if val, ok := data[key]; ok {

		if num, ok := val.(int); ok {
			return num
		}
	}
	return 0
}

func getFloat(data gin.H, key string) float64 {
	if val, ok := data[key]; ok {
		if num, ok := val.(float64); ok {
			return num
		}
	}
	return 0.0
}

func getString(data gin.H, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func convertToInterfaceSlice(strSlice []string) []interface{} {
	interfaceSlice := make([]interface{}, len(strSlice))
	for i, v := range strSlice {
		interfaceSlice[i] = v
	}
	return interfaceSlice
}
