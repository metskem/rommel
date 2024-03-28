package db

import (
	"database/sql"
	"errors"
	"github.com/metskem/rommel/joeradiostats/model"
	"log"
	"time"
)

// InsertSong - Insert a row into the song table. Returns the lastInsertId of the insert operation. */
func InsertSong(song model.Song) int64 {
	insertSQL := "insert into song(artist, title) values(?,?)"
	if statement, err := Database.Prepare(insertSQL); err != nil {
		log.Printf("failed to prepare stmt for insert into song, error: %s", err)
		return 0
	} else {
		defer func() { _ = statement.Close() }()
		if result, err := statement.Exec(song.Artist, song.Title); err != nil {
			log.Printf("failed to insert song %d, artist: %s, title %s, error: %s", song.Id, song.Artist, song.Title, err)
			return 0
		} else {
			if lastInsertId, err := result.LastInsertId(); err == nil {
				return lastInsertId
			} else {
				log.Printf("no song row was inserted, err: %s", err)
				return 0
			}
		}
	}
}

func GetSong(artist, title string) (model.Song, error) {
	var song model.Song
	selectSQL := "select s.id,s.artist,s.title,p.timestamp from song s, playmoment p where s.id=p.songid and artist=? and title=? order by p.timestamp desc limit 1"
	if statement, err := Database.Prepare(selectSQL); err != nil {
		return song, err
	} else {
		defer func() { _ = statement.Close() }()
		var id int64
		var lastplayed time.Time
		if err = statement.QueryRow(artist, title).Scan(&id, &artist, &title, &lastplayed); err != nil {
			return model.Song{}, err
		}
		return model.Song{Id: id, Artist: artist, Title: title, LastPlayed: lastplayed}, err
	}
}

func GetTotals() (int64, int64, int64, error) {
	selectSQL1 := "select count(*) from song s, playmoment p where s.id=p.songid;"
	selectSQL2 := "select count(*) from song"
	selectSQL3 := "select count(distinct(artist)) from song;select count(*) from song s, playmoment p where s.id=p.songid;"
	var cnt1, cnt2, cnt3 int64
	statement, _ := Database.Prepare(selectSQL1)
	defer func() { _ = statement.Close() }()
	if err := statement.QueryRow().Scan(&cnt1); err != nil {
		return 0, 0, 0, err
	}

	statement, _ = Database.Prepare(selectSQL2)
	defer func() { _ = statement.Close() }()
	if err := statement.QueryRow().Scan(&cnt2); err != nil {
		return 0, 0, 0, err
	}

	statement, _ = Database.Prepare(selectSQL3)
	defer func() { _ = statement.Close() }()
	if err := statement.QueryRow().Scan(&cnt3); err != nil {
		return 0, 0, 0, err
	}

	return cnt1, cnt2, cnt3, nil
}

func GetTopArtistsMostSongs() ([]model.ResultRow1, error) {
	var err error
	var rows *sql.Rows
	var result []model.ResultRow1
	queryString := "select artist,count(*) from song group by artist order by 2 desc,1 limit 10"
	if rows, err = Database.Query(queryString, nil); err != nil {
		return nil, err
	} else if rows == nil {
		return nil, errors.New("rows object was nil from GetTop10ArtistsMostSongs")
	} else {
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var artist string
			var count int
			err = rows.Scan(&artist, &count)
			if err != nil {
				log.Printf("error while scanning table: %s", err)
			}
			result = append(result, model.ResultRow1{Artist: artist, Count: count})
		}
	}
	return result, err
}

func GetTopArtistsMostOftenPlayed() ([]model.ResultRow1, error) {
	var err error
	var rows *sql.Rows
	var result []model.ResultRow1
	queryString := "select s.artist,count(*) from song s, playmoment p where s.id=p.songid group by artist order by 2 desc,1 limit 10"
	if rows, err = Database.Query(queryString, nil); err != nil {
		return nil, err
	} else if rows == nil {
		return nil, errors.New("rows object was nil from GetTop10ArtistsMostOftenPlayed")
	} else {
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var artist string
			var count int
			err = rows.Scan(&artist, &count)
			if err != nil {
				log.Printf("error while scanning table: %s", err)
			}
			result = append(result, model.ResultRow1{Artist: artist, Count: count})
		}
	}
	return result, err
}

func GetTopDuplicates() ([]model.ResultRow2, error) {
	var err error
	var rows *sql.Rows
	var result []model.ResultRow2
	queryString := "select s.artist,s.title,count(*) from song s, playmoment p where s.id=p.songid group by s.artist,s.title order by 3 desc limit 20"
	if rows, err = Database.Query(queryString, nil); err != nil {
		return nil, err
	} else if rows == nil {
		return nil, errors.New("rows object was nil from GetTopDuplicates")
	} else {
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var artist, title string
			var count int
			err = rows.Scan(&artist, &title, &count)
			if err != nil {
				log.Printf("error while scanning table: %s", err)
			}
			result = append(result, model.ResultRow2{Artist: artist, Title: title, Count: count})
		}
	}
	return result, err
}
