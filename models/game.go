package models

import "time"

type Game struct {
	ID            int       `db:"id" json:"id"`
	BotId         int       `db:"botid" json:"-"`
	Server        string    `db:"server" json:"-"`
	Map           string    `db:"map" json:"map"`
	DateTime      time.Time `db:"datetime" json:"datetime"`
	GameName      string    `db:"gamename" json:"-"`
	OwnerName     string    `db:"ownername" json:"-"`
	Duration      int       `db:"duration" json:"-"`
	GameState     int       `db:"gamestate" json:"-"`
	CreatorName   string    `db:"creatorname" json:"-"`
	CreatorServer string    `db:"creatorserver" json:"-"`
}
