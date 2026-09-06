package models

import (
	"testing"
	"time"
)

func ptrTime(t time.Time) *time.Time { return &t }

func testMatch() Match {
	start := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	end := start.Add(40 * time.Minute)

	return Match{
		Members: []Member{
			{MatchTeamID: 100, Private: MatchMemberPrivate{User: User{ID: 1, NickName: "alpha"}}},
			{MatchTeamID: 200, Private: MatchMemberPrivate{User: User{ID: 2, NickName: "beta"}}},
		},
		Maps: []MatchMap{
			{ID: 11, StartedAt: ptrTime(start), FinishedAt: ptrTime(end), Map: Map{Name: "dust2"}},
			{ID: 12, StartedAt: nil, FinishedAt: nil, Map: Map{Name: "unplayed"}}, // не сыграна
			{ID: 13, StartedAt: ptrTime(start), FinishedAt: ptrTime(end), Map: Map{Name: "mirage"}},
		},
		Teams: []MatchTeam{
			{ID: 100, MapStats: []MatchTeamMapStat{
				{MatchMapID: 11, Score: 13, IsWinner: true},
				{MatchMapID: 13, Score: 7, IsWinner: false},
			}},
			{ID: 200, MapStats: []MatchTeamMapStat{
				{MatchMapID: 11, Score: 9, IsWinner: false},
				{MatchMapID: 13, Score: 13, IsWinner: true},
			}},
		},
	}
}

// Несыгранная карта в середине списка раньше обрывала обработку
// всех последующих (break вместо continue).
func TestAddMatchDataContinuesPastUnplayedMap(t *testing.T) {
	st := &MatchApi{}
	st.AddMatchData(testMatch())

	if _, ok := st.Maps[11]; !ok {
		t.Error("карта 11 должна быть обработана")
	}
	if _, ok := st.Maps[12]; ok {
		t.Error("несыгранная карта 12 не должна попадать в Maps")
	}
	if _, ok := st.Maps[13]; !ok {
		t.Error("карта 13 после несыгранной должна быть обработана")
	}
}

// mapStats команд сопоставляются по MatchMapID, а не по позиции в срезе.
func TestAddMatchDataWinnerPerMap(t *testing.T) {
	st := &MatchApi{}
	st.AddMatchData(testMatch())

	if got := st.Maps[11].TeamWinnerId; got != 100 {
		t.Errorf("на карте 11 победитель 100, получено %d", got)
	}
	if got := st.Maps[13].TeamWinnerId; got != 200 {
		t.Errorf("на карте 13 победитель 200, получено %d", got)
	}
	if got := st.Maps[11].Rounds; got != 22 {
		t.Errorf("на карте 11 ожидалось 22 раунда, получено %d", got)
	}
}

// IsWinner не должен «залипать» с предыдущей карты.
func TestAddMatchDataIsWinnerResetPerMap(t *testing.T) {
	st := &MatchApi{}
	st.AddMatchData(testMatch())

	if !st.Maps[11].MapStats[1].IsWinner {
		t.Error("игрок 1 выиграл карту 11")
	}
	if st.Maps[13].MapStats[1].IsWinner {
		t.Error("игрок 1 проиграл карту 13, флаг должен сброситься")
	}
	if !st.Maps[13].MapStats[2].IsWinner {
		t.Error("игрок 2 выиграл карту 13")
	}
}

// Матч без двух команд не должен ронять обработку.
func TestAddMatchDataMissingTeams(t *testing.T) {
	match := testMatch()
	match.Teams = match.Teams[:1]

	st := &MatchApi{}
	st.AddMatchData(match)

	if len(st.Maps) != 0 {
		t.Errorf("без обеих команд карты не обрабатываются, получено %d", len(st.Maps))
	}
}
