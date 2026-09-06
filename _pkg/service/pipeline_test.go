package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	m "fastcup/_pkg/models"
)

// Сквозная проверка на реальном ответе Hasura: разбор -> AddMatchData ->
// ProcessStatistic -> CalculateDerivedStats. Раньше этот путь падал
// на декодировании (map.offset) и на nil-карте MapStats.
func TestFullPipelineOnRealMatch(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "models", "testdata", "match_response.json"))
	if err != nil {
		t.Skipf("нет образца ответа: %v", err)
	}

	var resp m.GetMatchStatsResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("ответ не разобрался: %v", err)
	}

	match := resp.Data.Match
	stats := m.MatchApi{}
	stats.AddMatchData(match)
	stats.InitPlayers(match.Members)

	if len(stats.Maps) == 0 {
		t.Fatal("не создано ни одной карты")
	}
	if len(stats.Players) == 0 {
		t.Fatal("не создано ни одного игрока")
	}

	// Без киллов/уронов Rounds всё равно должны быть проставлены из счёта команд
	for id, mp := range stats.Maps {
		if mp.MapStats == nil {
			t.Errorf("карта %d: MapStats == nil", id)
		}
		if mp.Rounds <= 0 {
			t.Errorf("карта %d: Rounds = %d", id, mp.Rounds)
		}
		if mp.MapName == "" {
			t.Errorf("карта %d: пустое имя", id)
		}
	}

	// Не должно паниковать даже без подгруженных киллов
	ProcessStatistic(&stats)

	// Производные метрики не должны давать NaN/Inf
	for _, mp := range stats.Maps {
		for pid, ps := range mp.MapStats {
			ps.Rounds = mp.Rounds
			ps.CalculateDerivedStats()
			if isBad(ps.Rating) || isBad(ps.Impact) || isBad(ps.KPR) || isBad(ps.DPR) {
				t.Errorf("игрок %d: некорректные метрики rating=%v impact=%v kpr=%v dpr=%v",
					pid, ps.Rating, ps.Impact, ps.KPR, ps.DPR)
			}
		}
	}
}

func isBad(f float64) bool {
	return f != f || f > 1e308 || f < -1e308
}
