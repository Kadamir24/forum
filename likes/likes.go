package likes

import (
	"database/sql"
	"fmt"
)

type Likes struct {
	DB *sql.DB
}

func (likes *Likes) GetOne(id, user string) Litem {
	item := Litem{}

	s := fmt.Sprintf("SELECT * FROM likes WHERE postid = '%v' AND username = '%v'", id, user)

	rows, _ := likes.DB.Query(s)
	var postid string
	var author string
	var like string
	if rows.Next() {
		rows.Scan(&postid, &author, &like)
		item = Litem{
			PostID:   postid,
			Username: author,
			Like:     like,
		}
	}
	rows.Close()
	return item
}

func (likes *Likes) Add(litem Litem) {
	one := likes.GetOne(litem.PostID, litem.Username)
	var s string
	if one.Like == "" {
		s = "INSERT INTO likes (like, postid, username) values (?, ?, ?)"
	} else if litem.Like != one.Like {
		s = "UPDATE likes SET like = ? WHERE postid = ? AND username = ?"
	} else {
		s = "DELETE FROM likes WHERE like = ? AND postid = ? AND username = ?"
	}
	stmt, _ := likes.DB.Prepare(s)
	_, err := stmt.Exec(litem.Like, litem.PostID, litem.Username)
	fmt.Println("this is likes db", err)
}

func (likes *Likes) Get(id, l string) []Litem {
	items := []Litem{}
	var s string
	if l == "all" {
		s = fmt.Sprintf("SELECT * FROM likes WHERE username = '%v'", id)

	} else {
		s = fmt.Sprintf("SELECT * FROM likes WHERE postid = '%v' AND like = '%v'", id, l)

	}

	rows, _ := likes.DB.Query(s)
	var postid string
	var author string
	var like string
	for rows.Next() {
		rows.Scan(&postid, &author, &like)
		item := Litem{
			PostID:   postid,
			Username: author,
			Like:     like,
		}
		items = append(items, item)
	}
	rows.Close()
	return items
}

func NewLD(db *sql.DB) *Likes {
	stmt, _ := db.Prepare(`
		CREATE TABLE IF NOT EXISTS "likes" (
			"postid"	TEXT NOT NULL,
			"username"	TEXT NOT NULL,
			"like"	TEXT
		);
	`)
	stmt.Exec()

	return &Likes{
		DB: db,
	}
}
