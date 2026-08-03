package model

// StatsResponse describes internal service statistics.
type StatsResponse struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}
