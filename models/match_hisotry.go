package models

import "time"

type MatchHistory struct {
	ID       int       `json:"id"`
	Map      string    `json:"map"`
	DateTime time.Time `json:"datetime"`
	Duration int       `json:"duration"`
	Teams    []Team    `json:"teams"`
}
