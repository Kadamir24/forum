package posts

import (
	"database/sql"
	"fmt"

	"01.alem.school/ilyasbulat/forum/likes"
)

type Post struct {
	DB *sql.DB
}

func (post *Post) GetMyPosts(LD *likes.Likes, str string) ([]Item, []Item) {
	s := fmt.Sprintf("SELECT * FROM posts WHERE author = '%v'", str)

	myitems := []Item{}
	mylikes := []Item{}
	likes := LD.Get(str, "all")
	rows, _ := post.DB.Query(s)
	var id string
	var author string
	var content string
	var thread string
	for rows.Next() {
		rows.Scan(&id, &author, &content, &thread)
		item := Item{
			ID:      id,
			Author:  author,
			Content: content,
			Thread:  thread,
			L:       len(LD.Get(id, "l")),
			D:       len(LD.Get(id, "d")),
		}
		myitems = append(myitems, item)
	}
	rows.Close()

	for _, v := range likes {
		s := fmt.Sprintf("SELECT * FROM posts WHERE id = '%v'", v.PostID)

		rows, _ := post.DB.Query(s)
		var id string
		var author string
		var content string
		var thread string
		var item Item
		if rows.Next() {
			rows.Scan(&id, &author, &content, &thread)
			item = Item{
				ID:      id,
				Author:  author,
				Content: content,
				Thread:  thread,
				L:       len(LD.Get(id, "l")),
				D:       len(LD.Get(id, "d")),
			}
			mylikes = append(mylikes, item)

		}

		rows.Close()
	}
	return myitems, mylikes
}

func (post *Post) Filter(LD *likes.Likes, str string) []Item {
	s := fmt.Sprintf("SELECT * FROM posts WHERE thread LIKE '%v'", "%"+str+"%")

	items := []Item{}
	rows, _ := post.DB.Query(s)
	var id string
	var author string
	var content string
	var thread string
	for rows.Next() {
		rows.Scan(&id, &author, &content, &thread)
		item := Item{
			ID:      id,
			Author:  author,
			Content: content,
			Thread:  thread,
			L:       len(LD.Get(id, "l")),
			D:       len(LD.Get(id, "d")),
		}
		items = append(items, item)
	}
	rows.Close()
	return items
}

func (post *Post) Get(LD *likes.Likes) []Item {
	items := []Item{}
	rows, _ := post.DB.Query(`
	SELECT * FROM "posts"
	`)
	var id string
	var author string
	var content string
	var thread string
	for rows.Next() {
		rows.Scan(&id, &author, &content, &thread)
		item := Item{
			ID:      id,
			Author:  author,
			Content: content,
			Thread:  thread,
			L:       len(LD.Get(id, "l")),
			D:       len(LD.Get(id, "d")),
		}
		items = append(items, item)
	}
	rows.Close()
	return items
}

// func (post *Post) GetOne(LD *likes.Likes, postid string) Item {
// 	item := Item{}

// 	s := fmt.Sprintf("SELECT * FROM posts WHERE id = '%v'", postid)

// 	rows, _ := post.DB.Query(s)
// 	var id string
// 	var author string
// 	var content string
// 	var thread string
// 	if rows.Next() {
// 		rows.Scan(&id, &author, &content, &thread)
// 		item = Item{
// 			ID:      id,
// 			Author:  author,
// 			Content: content,
// 			Thread:  thread,
// 			L:       len(LD.Get(id, "l")),
// 			D:       len(LD.Get(id, "d")),
// 		}
// 	}
// 	rows.Close()
// 	return item
// }

func (post *Post) Add(item Item) {
	fmt.Println("tyt", item)
	stmt, _ := post.DB.Prepare(`INSERT INTO "posts"(id, author, content, thread) values(?, ?, ?, ?)`)
	_, err := stmt.Exec(item.ID, item.Author, item.Content, item.Thread)
	fmt.Println(err)
}

func NewPost(db *sql.DB) *Post {
	stmt, _ := db.Prepare(`
		CREATE TABLE IF NOT EXISTS "posts" (
			"id"	TEXT NOT NULL UNIQUE,
			"author"	TEXT NOT NULL,
			"content"	TEXT NOT NULL,
			"thread"	TEXT,
			PRIMARY KEY("id")
		);
	`)
	stmt.Exec()

	return &Post{
		DB: db,
	}
}

func (post *Post) Update(item Item, id string) {
	stmt, _ := post.DB.Prepare(`UPDATE "posts" SET "content" = ?, "thread" = ? WHERE "id" = ?`)
	stmt.Exec(item.Content, item.Thread, id)
}

func (post *Post) Delete(id string) {
	stmt, _ := post.DB.Prepare(`DELETE FROM "posts" WHERE "id" = ?`)
	stmt.Exec(id)

}
