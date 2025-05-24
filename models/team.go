package models

type Team struct {
	Index    int       `json:"index"`
	Name     string    `json:"name"`
	Score    int       `json:"score"`
	Servants []Servant `json:"servants"`
}
