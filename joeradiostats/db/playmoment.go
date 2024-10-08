package db

import (
	"log"
	"time"
)

// InsertPlayMoment - Insert a row into the playmoment table. Returns the lastInsertId of the insert operation. */
func InsertPlayMoment(songId int64) int64 {
	insertSQL := "insert into playmoment(songid) values(?)"
	if statement, err := Database.Prepare(insertSQL); err != nil {
		log.Printf("failed to prepare stmt for insert into playmoment, error: %s", err)
		return 0
	} else {
		defer func() { _ = statement.Close() }()
		if result, err := statement.Exec(songId); err != nil {
			log.Printf("failed to insert playmoment for songid %d, error: %s", songId, err)
			return 0
		} else {
			if lastInsertId, err := result.LastInsertId(); err == nil {
				return lastInsertId
			} else {
				log.Printf("no playmoment row was inserted, err: %s", err)
				return 0
			}
		}
	}
}

func GetFirstPlayMoment() (time.Time, error) {
	selectSQL := "select timestamp from playmoment order by 1 limit 1;"
	var ts time.Time
	statement, _ := Database.Prepare(selectSQL)
	defer func() { _ = statement.Close() }()
	err := statement.QueryRow().Scan(&ts)
	return ts, err
}
