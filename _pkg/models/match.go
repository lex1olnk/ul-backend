package models

import "time"

type Match struct {
	ID                   int         `json:"id"`
	Type                 string      `json:"type"`
	Status               string      `json:"status"`
	BestOf               int         `json:"bestOf"`
	GameID               int         `json:"gameId"`
	HasWinner            bool        `json:"hasWinner"`
	StartedAt            time.Time   `json:"startedAt"`
	FinishedAt           *time.Time  `json:"finishedAt"`
	MaxRoundsCount       int         `json:"maxRoundsCount"`
	ServerInstanceID     *int        `json:"serverInstanceDd"`
	CancellationReason   *string     `json:"cancellation_reason"`
	ReplayExpirationDate *string     `json:"replayExpirationDate"`
	Rounds               []Round     `json:"rounds"`
	Maps                 []MatchMap  `json:"maps"`
	GameMode             GameMode    `json:"gameMode"`
	Teams                []MatchTeam `json:"teams"`
	Members              []Member    `json:"members"`
	Typename             string      `json:"__typename"`
}

type GameMode struct {
	ID             int    `json:"id"`
	TeamSize       int    `json:"teamSize"`
	FirstTeamSize  int    `json:"firstTeamSize"`
	SecondTeamSize int    `json:"secondTeamSize"`
	TypeName       string `json:"__typename"`
}

type Round struct {
	ID             int    `json:"id"`
	WinReason      string `json:"win_reason"`
	StartedAt      string `json:"startedAt"`
	FinishedAt     string `json:"finishedAt"`
	MatchMapID     int    `json:"matchMapId"`
	SpawnedPlayers []int  `json:"spawned_players"` // Исправлено на slice
	WinMatchTeamID *int   `json:"win_match_team_id"`
	Typename       string `json:"__typename"`
}

type MatchMap struct {
	ID         int              `json:"id"`
	Number     int              `json:"number"`
	MapID      int              `json:"mapId"`
	StartedAt  *time.Time       `json:"startedAt"`
	FinishedAt *time.Time       `json:"finishedAt"`
	GameStatus string           `json:"gameStatus"`
	Replays    []MatchMapReplay `json:"replays"`
	Map        Map              `json:"map"`
	Typename   string           `json:"__typename"`
}

type MatchMapReplay struct {
	ID        int    `json:"id"`
	URL       string `json:"url"`
	CreatedAt string `json:"createdAt"`
	TypeName  string `json:"__typename"`
}

type Map struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Offset   *float64 `json:"offset"` // Исправлено на указатель
	Scale    *float64 `json:"scale"`  // Исправлено на указатель
	Preview  string   `json:"preview"`
	Topview  string   `json:"topview"`
	Overview string   `json:"overview"`
	FlipV    bool     `json:"flip_v"`
	FlipH    bool     `json:"flip_h"`
	Typename string   `json:"__typename"`
}

type MatchTeam struct {
	ID             int                `json:"id"`
	Name           string             `json:"name"`
	Size           int                `json:"size"`
	Score          int                `json:"score"`
	ChatID         *int               `json:"chatId"` // Исправлено на *int
	IsWinner       bool               `json:"isWinner"`
	CaptainID      *int               `json:"captainId"` // Исправлено на *int
	IsDisqualified bool               `json:"isDisqualified"`
	MapStats       []MatchTeamMapStat `json:"mapStats"`
	Typename       string             `json:"__typename"`
}

type Member struct {
	Hash        string             `json:"hash"`
	Role        string             `json:"role"`
	Ready       bool               `json:"ready"`
	Impact      *float64           `json:"impact"` // Исправлено на указатель
	Connected   bool               `json:"connected"`
	IsLeaver    bool               `json:"isLeaver"`
	RatingDiff  *float64           `json:"ratingDiff"` // Исправлено на указатель
	MatchTeamID int                `json:"matchTeamId"`
	Private     MatchMemberPrivate `json:"private"`
	Typename    string             `json:"__typename"`
}

type MatchMemberPrivate struct {
	Rating   int    `json:"rating"`
	PartyID  int    `json:"partyId"`
	User     User   `json:"user"`
	TypeName string `json:"__typename"`
}

type User struct {
	ID                            int           `json:"id"`
	Link                          *string       `json:"link"` // Исправлено на указатель
	Avatar                        string        `json:"avatar"`
	Online                        bool          `json:"online"`
	Verified                      bool          `json:"verified"`
	IsMobile                      bool          `json:"isMobile"`
	NickName                      string        `json:"nickName"`
	AnimatedAvatar                *string       `json:"animatedAvatar"` // Исправлено на указатель
	IsMedia                       bool          `json:"isMedia"`
	DisplayMediaStatus            bool          `json:"displayMediaStatus"`
	PrivacyOnlineStatusVisibility string        `json:"privacyOnlineStatusVisibility"`
	Subscription                  *Subscription `json:"subscription"`
	Icon                          *ProfileIcon  `json:"icon"`
	Stats                         []UserStat    `json:"stats"`
	Typename                      string        `json:"__typename"`
}

type UserStat struct {
	Kills      int     `json:"kills"`
	Deaths     int     `json:"deaths"`
	Place      *int    `json:"place"` // Исправлено на указатель
	Rating     float64 `json:"rating"`
	WinRate    float64 `json:"winRate"`
	GameModeID int     `json:"gameModeId"`
	Typename   string  `json:"__typename"`
}

type MatchTeamMapStat struct {
	Score       int     `json:"score"`
	IsWinner    bool    `json:"isWinner"`
	MatchMapID  int     `json:"matchMapId"`
	MatchTeamID int     `json:"matchTeamId"`
	InitialSide *string `json:"initialSide"`
	TypeName    string  `json:"__typename"`
}

type Subscription struct {
	PlanID int `json:"planId"`
}

type ProfileIcon struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}
