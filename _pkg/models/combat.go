package models

import "time"

type Damage struct {
	RoundId          int    `json:"roundId"`
	InflictorId      int    `json:"inflictorId"`
	VictimId         int    `json:"victimId"`
	WeaponId         int    `json:"weaponId"`
	HitboxGroup      int    `json:"hitboxGroup"`
	DamageReal       int    `json:"damageReal"`
	DamageNormalized int    `json:"damageNormalized"`
	Hits             int    `json:"hits"`
	TypeName         string `json:"__typename"`
}

type Clutch struct {
	RoundId   int       `json:"roundId"`
	UserId    int       `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	Success   bool      `json:"success`
	Amount    int       `json:"amount"`
	Typename  string    `json:"__typename"`
}

type Kill struct {
	RoundId             int       `json:"roundId"`
	CreatedAt           time.Time `json:"createdAt"`
	KillerId            int       `json:"killerId"`
	VictimId            int       `json:"victimId"`
	AssistantId         *int      `json:"assistantId"` // Используем указатель, если может быть null
	WeaponId            int       `json:"weaponId"`
	IsHeadshot          bool      `json:"isHeadshot"`
	IsWallbang          bool      `json:"isWallbang"`
	IsOneshot           bool      `json:"isOneshot"`
	IsAirshot           bool      `json:"isAirshot"`
	IsNoscope           bool      `json:"isNoscope"`
	KillerPositionX     int       `json:"killerPositionX"`
	KillerPositionY     int       `json:"killerPositionY"`
	VictimPositionX     int       `json:"victimPositionX"`
	VictimPositionY     int       `json:"victimPositionY"`
	KillerBlindedBy     *int      `json:"killerBlindedBy"`     // Используем указатель, если может быть null
	KillerBlindDuration *int      `json:"killerBlindDuration"` // Используем указатель, если может быть null
	VictimBlindedBy     *int      `json:"victimBlindedBy"`
	VictimBlindDuration *int      `json:"victimBlindDuration"`
	IsTeamkill          bool      `json:"isTeamkill"`
	TypeName            string    `json:"__typename"`
}
