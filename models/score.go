package models

type Score struct {
	ID       int     `db:"id" json:"id"`
	Category string  `db:"category" json:"-"`
	Name     string  `db:"name" json:"name"`
	Server   string  `db:"server" json:"-"`
	Score    float64 `db:"score" json:"score"`
}
