package models

import "time"

type APIStats struct {
	MatchID    int       `json:"matchId"`
	Map        string    `json:"map"`
	Kills      int       `json:"kills"`
	Deaths     int       `json:"deaths"`
	Assists    int       `json:"assists"`
	Rounds     int       `json:"rounds"`
	Rating     float64   `json:"rating"`
	FinishedAt time.Time `json:"finishedAt"`
	// Другие приватные поля
}

type PlayerComparison struct {
	PlayerID          int     `json:"player_id"`
	Nickname          string  `json:"nickname"`
	ULRating          float64 `json:"ul_rating"`
	IMG               string  `json:"img"`
	Kills             int     `json:"kills"`
	Deaths            int     `json:"deaths"`
	Assists           int     `json:"assists"`
	FirstKills        int     `json:"fk"`
	FirstDeaths       int     `json:"fd"`
	KAST              float64 `json:"kast"`
	Maps              int     `json:"maps"`
	WinratePercentile float64 `json:"winrate_percentile"`
	KDPercentile      float64 `json:"kd_percentile"`
	HSPercentile      float64 `json:"hs_percentile"`
	AvgPercentile     float64 `json:"avg_percentile"`
	TargetWinrate     float64 `json:"target_winrate"`
	TargetKD          float64 `json:"target_kd"`
	TargetHSRatio     float64 `json:"target_hs_ratio"`
	TargetAvg         float64 `json:"target_avg"`
}

type MapsStatistic struct {
	Map       string
	Matches   int
	Wins      int
	AvgRating float64
	Winrate   float64
}
