package service

import (
	m "fastcup/_pkg/models"
)

type Mix struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Rating int    `json:"rating"`
}

type roundType struct {
	kills     int
	isDead    bool
	hasKill   bool
	hasTrade  bool
	hasAssist bool
}

func firstInteract(kill m.Kill) (int, int) {
	return kill.KillerId, kill.VictimId
}

func ProcessStatistic(st *m.MatchApi) {
	for _, round := range st.Rounds {
		entry := st.Maps[round.MapID]
		processRoundKills(round.Kills, &entry)
		processRoundDamages(round.Damages, &entry)
		processRoundClutches(round.Clutches, &entry)
		st.Maps[round.MapID] = entry
	}
}

func processRoundKills(kills []m.Kill, mapStat *m.Stats) {
	roundKAST := make(map[int]*roundType)
	fk, fd := firstInteract(kills[0])
	for index, kill := range kills {
		// Обновляем общую статистику
		updatePlayerMapStats(mapStat, kill)

		// Обрабатываем KAST и мультикиллы
		processKAST(roundKAST, kill)
		calculateTrade(index, kill, kills, mapStat)
	}

	// Финализируем статистику раунда
	finalizeRoundStats(mapStat, roundKAST, fk, fd)
}

func updatePlayerMapStats(mapStat *m.Stats, kill m.Kill) {
	// Киллер
	killerStats := mapStat.MapStats[kill.KillerId]
	killerStats.Kills++
	if kill.IsHeadshot {
		killerStats.Headshots++
	}
	mapStat.MapStats[kill.KillerId] = killerStats

	// Жертва
	victimStats := mapStat.MapStats[kill.VictimId]
	victimStats.Deaths++
	mapStat.MapStats[kill.VictimId] = victimStats

	if kill.VictimBlindedBy != nil {
		flashbanger := mapStat.MapStats[*kill.VictimBlindedBy]
		if flashbanger.TeamID == killerStats.TeamID {
			flashbanger.Flashes++
		}
		mapStat.MapStats[*kill.VictimBlindedBy] = flashbanger
	}

	// Ассистент
	if kill.AssistantId != nil {
		assistantStats := mapStat.MapStats[*kill.AssistantId]
		assistantStats.Assists++
		mapStat.MapStats[*kill.AssistantId] = assistantStats
	}

}

func processKAST(roundKAST map[int]*roundType, kill m.Kill) {
	// Инициализация записей игроков
	initPlayerKAST(roundKAST, kill.KillerId)
	initPlayerKAST(roundKAST, kill.VictimId)
	if kill.AssistantId != nil {
		initPlayerKAST(roundKAST, *kill.AssistantId)
	}

	// Обновление состояния
	roundKAST[kill.VictimId].isDead = true
	roundKAST[kill.KillerId].kills++
	roundKAST[kill.KillerId].hasKill = true

	if kill.AssistantId != nil {
		roundKAST[*kill.AssistantId].hasAssist = true
	}
}

func initPlayerKAST(roundKAST map[int]*roundType, playerID int) {
	if _, exists := roundKAST[playerID]; !exists {
		roundKAST[playerID] = &roundType{
			kills:     0,
			isDead:    false,
			hasKill:   false,
			hasTrade:  false,
			hasAssist: false,
		}
	}
}

func finalizeRoundStats(mapStat *m.Stats, roundKAST map[int]*roundType, fk int, fd int) {
	for playerID, stats := range roundKAST {
		playerStats := mapStat.MapStats[playerID]

		// Обновляем KAST
		if !stats.isDead || stats.hasKill || stats.hasAssist || stats.hasTrade {
			playerStats.KASTScore += 1.0
		}

		// Обновляем мультикиллы
		if stats.kills > 0 {
			idx := stats.kills - 1
			if idx >= 0 && idx < 5 {
				playerStats.MultiKills[idx]++
			}
		}
		if playerID == fk {
			playerStats.FirstKill++
		}
		if playerID == fd {
			playerStats.FirstDeath++
		}
		mapStat.MapStats[playerID] = playerStats
	}
}

func calculateTrade(index int, killer m.Kill, kills []m.Kill, stats *m.Stats) int {
	for i := index - 1; i > -1; i-- {
		difference := killer.CreatedAt.Sub(kills[i].CreatedAt)
		if difference.Seconds() > 5 {
			return 0
		}
		if killer.VictimId == kills[i].KillerId {
			kil := stats.MapStats[kills[i].VictimId]
			kil.Exchanged++
			stats.MapStats[kills[i].VictimId] = kil
			return kills[i].VictimId
		}
	}
	return 0
	//difference := trader.CreatedAt.Second() - killer.CreatedAt.Second()
	//fmt.Println(difference)
	// Время убийства

}

func processRoundDamages(damages []m.Damage, mapStat *m.Stats) {
	for _, damage := range damages {
		player := mapStat.MapStats[damage.InflictorId]
		player.Damage += float64(damage.DamageNormalized)
		if damage.WeaponId == 144 || damage.WeaponId == 142 || damage.WeaponId == 149 {
			player.NadesDamage += damage.DamageNormalized
		}
		mapStat.MapStats[damage.InflictorId] = player
	}
}

func processRoundClutches(clutches []m.Clutch, mapStat *m.Stats) {
	for _, clutch := range clutches {
		if clutch.Success {
			player := mapStat.MapStats[clutch.UserId]
			player.Clutches[clutch.Amount-1]++
			player.ClutchScore += clutch.Amount
			mapStat.MapStats[clutch.UserId] = player
		}
	}
}
