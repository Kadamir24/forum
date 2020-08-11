package comments

import (
	"database/sql"
	"fmt"

	"01.alem.school/ilyasbulat/forum/likescom"
)

type Comm struct {
	DB *sql.DB
}

func (comm *Comm) Delete(id string) {
	stmt, _ := comm.DB.Prepare(`DELETE FROM "comments" WHERE "commentid" = ?`)
	stmt.Exec(id)

}

func (comm *Comm) Get(LD *likescom.Likescom, str string) []Citem {
	s := fmt.Sprintf("SELECT * FROM comments WHERE postid = '%v'", str)

	items := []Citem{}
	rows, _ := comm.DB.Query(s)
	var commentid string
	var postid string
	var author string
	var content string
	for rows.Next() {
		rows.Scan(&commentid, &postid, &author, &content)
		item := Citem{
			CommentId: commentid,
			PostId:    postid,
			Author:    author,
			Content:   content,
			L:         len(LD.Get(commentid, "l")),
			D:         len(LD.Get(commentid, "d")),
		}
		items = append(items, item)
	}
	rows.Close()
	return items
}

func (comm *Comm) Add(citem Citem) {
	stmt, _ := comm.DB.Prepare(`INSERT INTO "comments" (commentid, postid, author, content) values(?, ?, ?, ?)`)
	_, err := stmt.Exec(citem.CommentId, citem.PostId, citem.Author, citem.Content)
	fmt.Println(citem)
	fmt.Println(err)
}

func NewComm(db *sql.DB) *Comm {
	stmt, _ := db.Prepare(`
		CREATE TABLE IF NOT EXISTS "comments" (
			"commentid" TEXT NOT NULL UNIQUE,
			"postid"	TEXT NOT NULL,
			"author"	TEXT NOT NULL,
			"content"	TEXT NOT NULL
		);
	`)
	stmt.Exec()
	return &Comm{
		DB: db,
	}
}
