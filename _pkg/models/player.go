package models

import "time"

const (
	kastWeight   = 0.008
	kprWeight    = 0.38
	dprWeight    = -0.576
	impactWeight = 0.28
	damageWeight = 0.0043
)

type MapStats struct {
	ID          int
	Nickname    string
	ULRating    float64
	IMG         string
	UL_ID       string
	Matches     int
	Kills       int
	Deaths      int
	Assists     int
	Headshots   int
	KASTScore   float64
	Damage      float64
	Exchanged   int
	FirstDeath  int
	FirstKill   int
	MultiKills  [5]int
	Clutches    [5]int
	Rounds      int
	TeamID      int
	KPR         float64
	DPR         float64
	Impact      float64
	ClutchScore int
	Rating      float64
	MatchID     int
	Map         string
	IsWinner    bool
	Flashes     int
	NadesDamage int
	NadeKill    int
	StartedAt   time.Time
	FinishedAt  time.Time
}

func (p *MapStats) CalculateDerivedStats() {
	p.KPR = float64(p.Kills) / float64(p.Rounds)
	p.DPR = float64(p.Deaths) / float64(p.Rounds)
	p.Impact = p.CalculateImpact()
	p.ClutchScore = p.CalculateClutchScore()
	p.Rating = p.CalculateRating()
}

func (p *MapStats) CalculateImpact() float64 {
	adrContribution := 0.2 * (p.Damage / 100)

	impact :=
		0.8*float64(p.FirstKill) +
			-0.6*float64(p.FirstDeath) + // Увеличил штраф
			0.72*float64(p.MultiKills[1]) + // 2K
			1.5*float64(p.MultiKills[2]) + // 3K
			2.0*float64(p.MultiKills[3]) + // 4K
			2.5*float64(p.MultiKills[4]) + // 5K
			0.8*float64(p.Clutches[0]) + // 1vs1
			1.25*float64(p.Clutches[1]) + // 1vs2
			1.6*float64(p.Clutches[2]) + // 1vs3
			2.2*float64(p.Clutches[3]) + // 1vs4
			3*float64(p.Clutches[4]) + // 1vs5
			adrContribution

	return impact / float64(p.Rounds)
}

func (p *MapStats) CalculateClutchScore() int {
	total := 0
	for _, c := range p.Clutches {
		total += c
	}
	return total
}

func (p *MapStats) CalculateRating() float64 {
	return kastWeight*p.KASTScore/float64(p.Rounds)*100 +
		kprWeight*(float64(p.Kills)/float64(p.Rounds)) +
		dprWeight*(float64(p.Deaths)/float64(p.Rounds)) +
		impactWeight*p.Impact +
		damageWeight*p.Damage/float64(p.Rounds) +
		0.24
}
