package model

import "time"

type Song struct {
	Id         int64
	Artist     string
	Title      string
	LastPlayed time.Time
}

type PlayMoment struct {
	Id        int64
	SongId    int64
	Timestamp time.Time
}

type ResultRow1 struct {
	Artist string
	Count  int
}

type ResultRow2 struct {
	Artist string
	Title  string
	Count  int
}
