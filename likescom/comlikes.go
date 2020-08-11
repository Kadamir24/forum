package likescom

import (
	"database/sql"
	"fmt"
)

type Likescom struct {
	DB *sql.DB
}

func (likes *Likescom) GetOne(id, user string) Litemcom {
	item := Litemcom{}

	s := fmt.Sprintf("SELECT * FROM likescom WHERE comid = '%v' AND username = '%v'", id, user)

	rows, _ := likes.DB.Query(s)
	var comid string
	var author string
	var like string
	if rows.Next() {
		rows.Scan(&comid, &author, &like)
		item = Litemcom{
			ComID:    comid,
			Username: author,
			Like:     like,
		}
	}
	rows.Close()
	return item
}

func (likes *Likescom) Add(litemcom Litemcom) {
	one := likes.GetOne(litemcom.ComID, litemcom.Username)
	var s string
	if one.Like == "" {
		s = "INSERT INTO likescom (like, comid, username) values (?, ?, ?)"
	} else if litemcom.Like != one.Like {
		s = "UPDATE likescom SET like = ? WHERE comid = ? AND username = ?"
	} else {
		s = "DELETE FROM likescom WHERE like = ? AND comid = ? AND username = ?"
	}
	stmt, _ := likes.DB.Prepare(s)
	_, err := stmt.Exec(litemcom.Like, litemcom.ComID, litemcom.Username)
	fmt.Println("this is likescom db", err)
}

func (likes *Likescom) Get(id, l string) []Litemcom {
	items := []Litemcom{}
	var s string
	if l == "all" {
		s = fmt.Sprintf("SELECT * FROM likescom WHERE username = '%v'", id)

	} else {
		s = fmt.Sprintf("SELECT * FROM likescom WHERE comid = '%v' AND like = '%v'", id, l)

	}

	rows, _ := likes.DB.Query(s)
	var postid string
	var author string
	var like string
	for rows.Next() {
		rows.Scan(&postid, &author, &like)
		item := Litemcom{
			ComID:    postid,
			Username: author,
			Like:     like,
		}
		items = append(items, item)
	}
	rows.Close()
	return items
}

func NewLDcom(db *sql.DB) *Likescom {
	stmt, _ := db.Prepare(`
		CREATE TABLE IF NOT EXISTS "likescom" (
			"comid"	TEXT NOT NULL,
			"username"	TEXT NOT NULL,
			"like"	TEXT
		);
	`)
	stmt.Exec()

	return &Likescom{
		DB: db,
	}
}
