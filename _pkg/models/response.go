package models

type GraphQLClutchResponse struct {
	Data struct {
		Clutches []Clutch `json:"clutches"`
	} `json:"data"`
}

type GraphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]int `json:"variables"`
}

type GraphQLKillsResponse struct {
	Data struct {
		Kills []Kill `json:"kills"`
	} `json:"data"`
}

type GraphQLDamagesResponse struct {
	Data struct {
		Damages []Damage `json:"damages"`
	} `json:"data"`
}

type GetMatchStatsResponse struct {
	Data struct {
		Match Match `json:"match"`
	} `json:"data"`
}
