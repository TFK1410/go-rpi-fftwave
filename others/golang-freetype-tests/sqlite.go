package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

func (ldc *LyricDrawContext) startSqlite() {
	var err error
	ldc.db, err = sql.Open("sqlite3", "./lyrics.sqlite")
	if err != nil {
		log.Fatal(err)
	}
}

func (ldc *LyricDrawContext) getLyric(id int) *Lyric {
	sqlStatement := `SELECT * FROM Lyrics WHERE LyricID=$1;`
	var lyricid int
	var trackid int
	var lyric string
	var attributes string

	row := ldc.db.QueryRow(sqlStatement, id)
	switch err := row.Scan(&lyricid, &trackid, &lyric, &attributes); err {
	case sql.ErrNoRows:
		log.Println("No rows were returned!")
	case nil:
		l := NewLyric()
		l.parseLyricData(lyricid, lyric, attributes, ldc)
		return l
	default:
		log.Fatal(err)
	}

	return nil
}
