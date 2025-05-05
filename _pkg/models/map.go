package models

import (
	"time"
)

type PlayerInit struct {
	ID       int
	Nickname string
	ULRating float64
	IMG      string
}

type Stats struct {
	MapID        int
	MapName      string
	Rounds       int
	TeamWinnerId int
	StartedAt    time.Time
	FinishedAt   time.Time
	MapStats     map[int]MapStats
}

type GraphQlRound struct {
	MapID    int
	MapName  string
	Kills    []Kill
	Clutches []Clutch
	Damages  []Damage
}

type MatchApi struct {
	MatchId    int
	Players    []PlayerInit
	StartedAt  time.Time
	FinishedAt time.Time
	WinnerId   int
	Maps       map[int]Stats
	Rounds     map[int]GraphQlRound
}

func (st *MatchApi) AddMatchData(match Match) {
	players := map[int]MapStats{}

	for _, player := range match.Members {
		user := player.Private.User
		if _, ok := players[user.ID]; !ok {
			p := MapStats{}
			p.ID = user.ID
			p.Nickname = user.NickName
			p.TeamID = player.MatchTeamID
			p.IMG = user.Avatar
			players[user.ID] = p
			p.MultiKills = [5]int{0, 0, 0, 0, 0}
			p.Clutches = [5]int{0, 0, 0, 0, 0}
		}
	}

	st.Maps = map[int]Stats{}

	for index, m := range match.Maps {
		if m.StartedAt == nil {
			break
		}
		winner := match.Teams[0].ID
		if !match.Teams[0].MapStats[index].IsWinner {
			winner = match.Teams[1].ID
		}

		for id, player := range players {
			if player.TeamID == winner {
				player.IsWinner = true
			}
			player.StartedAt = *m.StartedAt
			player.FinishedAt = *m.FinishedAt
			players[id] = player
		}

		rounds := match.Teams[0].MapStats[index].Score + match.Teams[1].MapStats[index].Score
		newmap := Stats{
			m.ID,
			m.Map.Name,
			rounds,
			winner,
			*m.StartedAt,
			*m.FinishedAt,
			players,
		}
		st.Maps[m.ID] = newmap
	}
	st.Rounds = map[int]GraphQlRound{}
	for _, r := range match.Rounds {
		if _, ok := st.Rounds[r.ID]; !ok {
			st.Rounds[r.ID] = GraphQlRound{
				r.MatchMapID,
				"unknown",
				[]Kill{},
				[]Clutch{},
				[]Damage{},
			}
		}
	}
}

func (st *MatchApi) InitPlayers(members []Member) {
	st.Players = make([]PlayerInit, 0)

	for _, member := range members {
		st.Players = append(st.Players, PlayerInit{
			member.Private.User.ID,
			member.Private.User.NickName,
			0,
			member.Private.User.Avatar,
		})
	}
}

func (st *MatchApi) InitKills(kills []Kill) {
	for _, kill := range kills {
		entry := st.Rounds[kill.RoundId]
		entry.Kills = append(entry.Kills, kill)
		st.Rounds[kill.RoundId] = entry
	}
}

func (st *MatchApi) InitClutch(clutches []Clutch) {
	for _, clutch := range clutches {
		entry := st.Rounds[clutch.RoundId]
		entry.Clutches = append(entry.Clutches, clutch)
		st.Rounds[clutch.RoundId] = entry
	}
}

func (st *MatchApi) InitDamage(damages []Damage) {
	for _, damage := range damages {
		entry := st.Rounds[damage.RoundId]
		entry.Damages = append(entry.Damages, damage)
		st.Rounds[damage.RoundId] = entry
	}
}

/*
func (st *MatchApi) arrangeKills(match Match) {
	for _, round := range match.Rounds {
		// Создаем или находим статистику для карты
		// Обрабатываем киллы для текущего раунда
		s.processRoundKills(round.Kills, mapStat, round.MapID)
	}
}

func (st *MatchApi) arrangeClutches(clutches []Clutch) {
	for _, clutch := range clutches {
		if clutch.Success {
			stats[clutch.RoundId].Clutches = append(st[clutch.RoundId].Clutches, clutch)
		}
	}
}
*/
