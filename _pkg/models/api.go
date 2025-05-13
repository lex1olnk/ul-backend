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
	IsWinner   bool      `json:"isWinner"`
	FinishedAt time.Time `json:"finishedAt"`
	// Другие приватные поля
}

type PlayerComparison struct {
	PlayerID         int     `json:"player_id"`
	Nickname         string  `json:"nickname"`
	Faceit           string  `json:"faceit"`
	ULRating         float64 `json:"ul_rating"`
	IMG              string  `json:"img"`
	Kills            int     `json:"kills"`
	Deaths           int     `json:"deaths"`
	Assists          int     `json:"assists"`
	FirstKills       int     `json:"fk"`
	FirstDeaths      int     `json:"fd"`
	Flashes          int     `json:"flashes`
	Nades            int     `json:"nades`
	Impact           float64 `json:"impact`
	KAST             float64 `json:"kast"`
	Maps             int     `json:"maps"`
	Exchanged        int     `json:"exchanged"`
	V1               int     `json:"clutches_1v1"`
	V2               int     `json:"clutches_1v2"`
	V3               int     `json:"clutches_1v3"`
	V4               int     `json:"clutches_1v4"`
	V5               int     `json:"clutches_1v5"`
	Winrate          float64 `json:"winrate"`
	RatingPercentile float64 `json:"rating_percentile"`
	KDPercentile     float64 `json:"kd_percentile"`
	HSPercentile     float64 `json:"hs_percentile"`
	AvgPercentile    float64 `json:"avg_percentile"`
	TargetKD         float64 `json:"target_kd"`
	TargetHSRatio    float64 `json:"target_hs_ratio"`
	TargetAvg        float64 `json:"target_avg"`
	Rating           float64 `json:"target_rating"`
}

type MapsStatistic struct {
	Map       string
	Matches   int
	Wins      int
	AvgRating float64
	Winrate   float64
}
