package service

import (
	"testing"
	"time"

	m "fastcup/_pkg/models"
)

// Раунд, ссылающийся на карту, которой нет в st.Maps, раньше приводил
// к записи в nil-карту MapStats и панике.
func TestProcessStatisticSkipsUnknownMap(t *testing.T) {
	st := &m.MatchApi{
		Players: []m.PlayerInit{{ID: 1}, {ID: 2}},
		Maps:    map[int]m.Stats{},
		Rounds: map[int]m.GraphQlRound{
			10: {
				MapID: 999, // такой карты в Maps нет
				Kills: []m.Kill{{KillerId: 1, VictimId: 2, CreatedAt: time.Now()}},
			},
		},
	}

	ProcessStatistic(st) // не должно паниковать
}

// Карта присутствует, но MapStats == nil — тоже не повод падать.
func TestProcessStatisticSkipsNilMapStats(t *testing.T) {
	st := &m.MatchApi{
		Players: []m.PlayerInit{{ID: 1}},
		Maps:    map[int]m.Stats{5: {MapID: 5, MapStats: nil}},
		Rounds: map[int]m.GraphQlRound{
			1: {MapID: 5, Kills: []m.Kill{{KillerId: 1, VictimId: 2, CreatedAt: time.Now()}}},
		},
	}

	ProcessStatistic(st)
}

// Раунд без убийств раньше падал на kills[0].
func TestProcessRoundKillsEmpty(t *testing.T) {
	mapStat := m.Stats{MapStats: map[int]m.MapStats{}}
	roundKAST := map[int]*roundType{}

	processRoundKills(nil, &mapStat, roundKAST)
}

// Убийца, которого нет среди участников матча, раньше давал nil-указатель.
func TestProcessKASTUnknownPlayer(t *testing.T) {
	roundKAST := map[int]*roundType{}
	assistant := 77

	processKAST(roundKAST, m.Kill{KillerId: 42, VictimId: 43, AssistantId: &assistant})

	for _, id := range []int{42, 43, 77} {
		if roundKAST[id] == nil {
			t.Fatalf("ожидалась запись KAST для игрока %d", id)
		}
	}
	if !roundKAST[42].hasKill {
		t.Error("у убийцы должен быть выставлен hasKill")
	}
	if !roundKAST[43].isDead {
		t.Error("у жертвы должен быть выставлен isDead")
	}
	if !roundKAST[77].hasAssist {
		t.Error("у ассистента должен быть выставлен hasAssist")
	}
}

// Amount вне диапазона 1..5 раньше выходил за границы массива Clutches.
func TestProcessRoundClutchesOutOfRange(t *testing.T) {
	mapStat := m.Stats{MapStats: map[int]m.MapStats{1: {ID: 1}}}

	processRoundClutches([]m.Clutch{
		{UserId: 1, Success: true, Amount: 0},
		{UserId: 1, Success: true, Amount: 6},
		{UserId: 1, Success: true, Amount: 2}, // валидный
	}, &mapStat)

	if got := mapStat.MapStats[1].Clutches[1]; got != 1 {
		t.Errorf("ожидался один клатч 1v2, получено %d", got)
	}
	if got := mapStat.MapStats[1].ClutchScore; got != 2 {
		t.Errorf("ожидался ClutchScore 2, получено %d", got)
	}
}
