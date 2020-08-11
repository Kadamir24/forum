package users

import (
	"database/sql"
	"fmt"
)

type User struct {
	DB *sql.DB
}

// need to return error and handle it on server

func (user *User) Add(uitem Uitem) error {
	stmt, _ := user.DB.Prepare(`
	INSERT INTO "users" (email, username, password) values (?, ?, ?)
	`)
	_, err := stmt.Exec(uitem.Email, uitem.Username, uitem.Password)
	if err != nil {
		return err
	}
	return nil
}

func NewUser(db *sql.DB) *User {
	stmt, _ := db.Prepare(`
		CREATE TABLE IF NOT EXISTS "users" (
			"email"	TEXT UNIQUE,
			"username"	TEXT NOT NULL UNIQUE,
			"password"	TEXT NOT NULL,
			PRIMARY KEY("username")
		);
	`)
	stmt.Exec()

	return &User{
		DB: db,
	}
}

func (user *User) Get() []Uitem {
	uitems := []Uitem{}
	rows, _ := user.DB.Query(`
	SELECT * FROM "users"
	`)
	var email string
	var username string
	var password string
	for rows.Next() {
		rows.Scan(&email, &username, &password)
		uitem := Uitem{
			Email:    email,
			Username: username,
			Password: password,
		}
		uitems = append(uitems, uitem)
	}
	rows.Close()
	return uitems
}

func (user *User) GetUser(str string) Uitem {
	s := fmt.Sprintf("SELECT * FROM users WHERE username = '%v'", str)
	rows, _ := user.DB.Query(s)
	var email string
	var username string
	var pass string
	var uitem Uitem
	if rows.Next() {
		rows.Scan(&email, &username, &pass)
		uitem = Uitem{
			Email:    email,
			Username: username,
			Password: pass,
		}
	}
	rows.Close()
	return uitem
}
