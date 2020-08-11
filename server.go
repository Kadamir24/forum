package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"

	"01.alem.school/ilyasbulat/forum/comments"
	"01.alem.school/ilyasbulat/forum/likes"
	"01.alem.school/ilyasbulat/forum/likescom"
	"01.alem.school/ilyasbulat/forum/posts"
	"01.alem.school/ilyasbulat/forum/users"
	_ "github.com/mattn/go-sqlite3"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

var bunch *posts.Post
var guys *users.User
var flood *comments.Comm
var templates *template.Template
var LD *likes.Likes
var LDcom *likescom.Likescom

func main() {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("template"))
	mux.Handle("/template/", http.StripPrefix("/template/", fs))

	db, _ := sql.Open("sqlite3", "./posts.db")
	defer db.Close()
	guys = users.NewUser(db)
	bunch = posts.NewPost(db)
	flood = comments.NewComm(db)
	LD = likes.NewLD(db)
	LDcom = likescom.NewLDcom(db)
	mux.HandleFunc("/", Middleware(Postsss))

	// mux.HandleFunc("/posts", Middleware(Postsss))
	mux.HandleFunc("/write", Middleware(newPost))
	mux.HandleFunc("/edit", Middleware(newPost))
	mux.HandleFunc("/view", Middleware(View))
	mux.HandleFunc("/comment", Middleware(View))
	mux.HandleFunc("/delete", Middleware(deletePost))
	mux.HandleFunc("/SavePost", Middleware(savePost))
	mux.HandleFunc("/savecomm", Middleware(SaveComm))
	mux.HandleFunc("/login", Middleware(login))
	mux.HandleFunc("/register", Middleware(register))
	mux.HandleFunc("/deleteCom", Middleware(DelComm))
	mux.HandleFunc("/like", Middleware(LikeDislike))
	mux.HandleFunc("/dislike", Middleware(LikeDislike))
	mux.HandleFunc("/logout", Middleware(Logout))
	mux.HandleFunc("/filter", Middleware(Filter))
	mux.HandleFunc("/cabinet", Middleware(Cabinet))
	mux.HandleFunc("/likecom", Middleware(LikeDislikecom))
	mux.HandleFunc("/dislikecom", Middleware(LikeDislikecom))
	e := http.ListenAndServe(":8080", mux)
	if e != nil {
		fmt.Println(e)
	}
}

func LikeDislikecom(w http.ResponseWriter, r *http.Request, s *Session) {
	if !s.IsAuthorized {
		http.Redirect(w, r, "/", 302)
		return
	}
	var value string
	if r.URL.Path == "/likecom" {
		value = "l"
	} else if r.URL.Path == "/dislikecom" {
		value = "d"
	}
	values, _ := url.ParseQuery(r.URL.RawQuery)
	comid := values.Get("coid")
	posid := values.Get("posid")
	LDcom.Add(likescom.Litemcom{
		ComID:    comid,
		Username: s.Username,
		Like:     value,
	})
	http.Redirect(w, r, "/view?id="+posid, 302)
}

func Cabinet(w http.ResponseWriter, r *http.Request, s *Session) {
	myposts, mylikes := bunch.GetMyPosts(LD, s.Username)
	data := Info{
		Sess:       s,
		Posts:      myposts,
		LikedPosts: mylikes,
	}
	t, err := template.ParseFiles("template/cabinet.html", "template/header.html", "template/footer.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Println(mylikes)
	t.ExecuteTemplate(w, "cabinet", data)

}

func Filter(w http.ResponseWriter, r *http.Request, s *Session) {
	thread := r.FormValue("thread")
	p := bunch.Filter(LD, thread)
	data := Info{
		Sess:  s,
		Posts: p,
	}
	t, err := template.ParseFiles("template/posts.html", "template/header.html", "template/footer.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	t.ExecuteTemplate(w, "posts", data)
}

func Logout(w http.ResponseWriter, r *http.Request, s *Session) {
	sessionStore.Delete(s)
	http.Redirect(w, r, "/", 302)
}

func LikeDislike(w http.ResponseWriter, r *http.Request, s *Session) {
	if !s.IsAuthorized {
		http.Redirect(w, r, "/", 302)
		return
	}
	var value string
	if r.URL.Path == "/like" {
		value = "l"
	} else if r.URL.Path == "/dislike" {
		value = "d"
	}
	LD.Add(likes.Litem{
		PostID:   r.FormValue("id"),
		Username: s.Username,
		Like:     value,
	})
	http.Redirect(w, r, "/", 302)
}

func DelComm(w http.ResponseWriter, r *http.Request, s *Session) {
	// id := strings.TrimPrefix(r.URL.RequestURI(), "/deleteCom?")
	// ids := strings.Split(r.URL.RequestURI(), "?")
	values, _ := url.ParseQuery(r.URL.RawQuery)
	comid := values.Get("coid")
	posid := values.Get("posid")
	flood.Delete(comid)
	http.Redirect(w, r, "/view?id="+posid, 302)

	// if len(ids) == 3 {
	// 	comid := ids[1]
	// 	postid := ids[2]
	// 	flood.Delete(comid)
	// http.Redirect(w, r, "/view?"+postid, 302)
	// 	return
	// }
	// http.Redirect(w, r, "/view", 500)

}

func SaveComm(w http.ResponseWriter, r *http.Request, s *Session) {
	if r.FormValue("content") != "" && verifyContent(r.FormValue("content")) {
		flood.Add(comments.Citem{
			CommentId: Generate(),
			PostId:    r.FormValue("id"),
			Author:    s.Username,
			Content:   r.FormValue("content"),
		})
	}

	http.Redirect(w, r, "/view?id="+r.FormValue("id"), 302)
}

func View(w http.ResponseWriter, r *http.Request, s *Session) {
	var url string
	if r.URL.Path != "/view" && s.Username != "" {

		url = "comment"
	} else {
		url = "view"
	}

	items := bunch.Get(LD)
	// id := strings.TrimPrefix(r.URL.RequestURI(), "/"+url+"?")
	id := r.FormValue("id")
	var item posts.Item

	for _, v := range items {
		if v.ID == id {
			item = v
		}
	}
	coms := flood.Get(LDcom, id)
	for i, v := range coms {
		if v.Author == s.Username {
			coms[i].ComIsAuthor = true
		}
	}

	data := Info{
		Sess:     s,
		Comments: coms,
		Post:     item,
		IsAuthor: item.Author == s.Username,
	}

	t, err := template.ParseFiles("template/comment.html", "template/view.html", "template/header.html", "template/footer.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	t.ExecuteTemplate(w, url, data)
}

func register(w http.ResponseWriter, r *http.Request, s *Session) {
	t, err := template.ParseFiles("template/auth.html", "template/header.html", "template/footer.html")
	if err != nil {
		// 500 return
		return
		// panic(err)
	}

	if r.Method == "GET" {

		e := t.ExecuteTemplate(w, "register", nil)
		if e != nil {
			fmt.Fprint(w, e)
		}
	} else {
		//registration logic
		type Data struct {
			Email      string
			Username   string
			Password   string
			PHemail    string
			PHusername string
		}

		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")
		data := Data{
			PHemail:    email,
			PHusername: username,
		}
		if m, _ := regexp.MatchString(`^([\w\.\_]{2,10})@(\w{1,}).([a-z]{2,4})$`, email); !m {
			data.Email = "need an apropriate email address"
			t.ExecuteTemplate(w, "register", data)

			return
		}

		if guys.GetUser(username).Username == username {
			data.Username = "already taken"
			t.ExecuteTemplate(w, "register", data)
			return
		}
		if m, _ := regexp.MatchString(`^\w{5,10}$`, username); !m {
			data.Username = "only eng letters from 5 to 10 symbols, no spaces"
			t.ExecuteTemplate(w, "register", data)
			return
		}

		if m := verifyPassword(password); !m {
			fmt.Println(m)
			data.Password = "must be eng letters at least 1 Upper case and special symbol, 6 to 10 long, no spaces"
			err := t.ExecuteTemplate(w, "register", data)
			if err != nil {
				http.Error(w, err.Error(), 500)
			}
			return
		}

		hashedPass, err := HashPassword(password)
		if err != nil {
			log.Fatal(err.Error())
		}

		err1 := guys.Add(users.Uitem{
			Email:    email,
			Username: username,
			Password: hashedPass,
		})
		if err1 != nil {
			data.Email = "user with this email already exists"
			e := t.ExecuteTemplate(w, "register", data)
			if e != nil {
				http.Error(w, e.Error(), 500)
			}
			return
		}

		http.Redirect(w, r, "/login", 302)
	}

}

func login(w http.ResponseWriter, r *http.Request, s *Session) {
	if AlreadyLoggedIn(r) {
		http.Redirect(w, r, "/", 302)
	}
	t, err := template.ParseFiles("template/login.html", "template/header.html", "template/footer.html")
	if err != nil {

		return
		// panic(err)
	}

	if r.Method == "GET" {

		e := t.ExecuteTemplate(w, "login", nil)
		if e != nil {
			fmt.Fprint(w, e)
		}
	} else {
		//login logic

		type Data struct {
			Username string
			Password string
		}

		var data Data

		username := r.FormValue("username")

		password := r.FormValue("password")
		user := guys.GetUser(username)

		if user.Username == username {
			if CheckPasswordHash(password, user.Password) {
				for k, v := range sessionStore.data {
					if v.Username == username {
						delete(sessionStore.data, k)
					}
				}
				s.IsAuthorized = true
				s.Username = username

				http.Redirect(w, r, "/", 302)
				return

			}
			data.Password = "this password is incorrect"
			t.ExecuteTemplate(w, "login", data)

			return
		}
		data.Username = "no such user"
		t.ExecuteTemplate(w, "login", data)

	}

}

// func indexHandler(w http.ResponseWriter, r *http.Request, s *Session) {

// 	if r.URL.Path != "/" {
// 		w.WriteHeader(404)
// 		w.Write([]byte("not found"))
// 		return
// 	}
// 	t, err := template.ParseFiles("template/index.html", "template/header.html", "template/footer.html")
// 	if err != nil {
// 		fmt.Fprintf(w, err.Error())
// 		return
// 		// panic(err)
// 	}

// 	e := t.ExecuteTemplate(w, "index", s)
// 	if e != nil {
// 		fmt.Fprint(w, e)
// 	}

// }

func deletePost(w http.ResponseWriter, r *http.Request, s *Session) {

	id := strings.TrimPrefix(r.URL.RequestURI(), "/delete?")
	bunch.Delete(id)
	http.Redirect(w, r, "/", 302)

}

func savePost(w http.ResponseWriter, r *http.Request, s *Session) {
	if s.Username == "" {
		http.Redirect(w, r, "/", 302)
	}

	fmt.Println(r.FormValue("content"), verifyContent(r.FormValue("content")))
	if r.FormValue("content") != "" && verifyContent(r.FormValue("content")) {
		if r.FormValue("id") != "" {
			id := r.FormValue("id")
			r.ParseForm()
			var thread string
			threadList := r.Form["thread"]
			if len(threadList) > 1 {
				for i, v := range threadList {
					thread += v
					if i != len(threadList)-1 {
						thread += ":"
					}

				}
			} else {
				thread = threadList[0]
			}
			bunch.Update(posts.Item{
				Content: r.FormValue("content"),
				Thread:  thread,
			}, id)
		} else {
			r.ParseForm()
			var thread string
			threadList := r.Form["thread"]
			if len(threadList) > 1 {
				for i, v := range threadList {
					thread += v
					if i != len(threadList)-1 {
						thread += ":"
					}

				}
			} else {
				thread = threadList[0]
			}

			bunch.Add(posts.Item{
				ID:      Generate(),
				Author:  s.Username,
				Content: r.FormValue("content"),
				Thread:  thread,
			})

		}

		http.Redirect(w, r, "/", 302)

	} else {
		t, err := template.ParseFiles("template/newpost.html", "template/header.html", "template/footer.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		data := Info{
			Sess:  s,
			Error: "You need to write something at least 8 symbols",
			Post: posts.Item{
				Content: r.FormValue("content"),
			},
		}
		t.ExecuteTemplate(w, "newpost", data)

	}

}

func newPost(w http.ResponseWriter, r *http.Request, s *Session) {
	if s.Username == "" {
		http.Redirect(w, r, "/", 302)
	}
	items := bunch.Get(LD)
	id := strings.TrimPrefix(r.URL.RequestURI(), "/edit?")
	var item posts.Item

	for _, v := range items {
		if v.ID == id {
			item = v
		}
	}
	t, err := template.ParseFiles("template/newpost.html", "template/header.html", "template/footer.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data := Info{
		Sess: s,
		Post: item,
	}
	t.ExecuteTemplate(w, "newpost", data)

}

type Info struct {
	Sess       *Session
	Comments   []comments.Citem
	Posts      []posts.Item
	LikedPosts []posts.Item
	Post       posts.Item
	IsAuthor   bool
	Error      string
}

func Postsss(w http.ResponseWriter, r *http.Request, s *Session) {
	if r.URL.Path != "/" {
		w.WriteHeader(404)
		w.Write([]byte("not found"))
		return
	}
	items := bunch.Get(LD)
	var data Info
	// for i, v := range items {
	// 	if v.Author == s.Username {
	// 		items[i].IsAuthor = true
	// 	}
	// }
	data = Info{
		Sess:  s,
		Posts: items,
	}
	t, err := template.ParseFiles("template/posts.html", "template/header.html", "template/footer.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	t.ExecuteTemplate(w, "posts", data)

}

//utilities

func Generate() string {
	u2 := uuid.NewV4()
	return fmt.Sprintf("%x", u2)
}

func parseHTMLFiles(file string) (*template.Template, error) {
	t, err := template.ParseFiles(file, "static/templates/header.html", "static/templates/footer.html")
	return t, err
}

func verifyPassword(s string) bool {
	letters := 0
	var sevenOrTen, number, upper, special bool
	for _, c := range s {
		switch {
		case unicode.IsNumber(c):
			number = true
			letters++
		case unicode.IsUpper(c):
			upper = true
			letters++
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			special = true
		case unicode.IsLetter(c):
			letters++
		default:
			//return false, false, false, false
		}
	}
	sevenOrTen = letters >= 8 && letters <= 10
	if sevenOrTen && number && upper && special {
		return true
	}
	return false
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

//sessions
const (
	COOKIE_NAME = "sessionId"
)

type Session struct {
	Id           string
	Username     string
	IsAuthorized bool
}

type SessionStore struct {
	data map[string]*Session
}

var sessionStore = NewSessionStore()

func NewSessionStore() *SessionStore {
	s := new(SessionStore)
	s.data = make(map[string]*Session)
	return s
}

func (store *SessionStore) Get(sessionId string) *Session {
	session := store.data[sessionId]
	if session == nil {
		return &Session{Id: sessionId}
	}
	return session
}

func (store *SessionStore) Set(session *Session) {
	store.data[session.Id] = session
}

func (store *SessionStore) Delete(session *Session) {
	delete(store.data, session.Id)
}

func ensureCookie(r *http.Request, w http.ResponseWriter) string {
	cookie, _ := r.Cookie(COOKIE_NAME)
	if cookie != nil {
		if cookie.Expires.Before(time.Now()) {
			cookie.Expires = time.Now().Add(5 * time.Minute)
			http.SetCookie(w, cookie)

		}
		return cookie.Value
	}
	sessionId := Generate()

	cookie = &http.Cookie{
		Name:    COOKIE_NAME,
		Value:   sessionId,
		Expires: time.Now().Add(5 * time.Minute),
	}
	http.SetCookie(w, cookie)

	return sessionId
}

func Middleware(next func(w http.ResponseWriter, r *http.Request, s *Session)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		sessionId := ensureCookie(r, w)

		session := sessionStore.Get(sessionId)

		sessionStore.Set(session)
		next(w, r, session)
	}
}

// not sure if that next func is nedded, cuz middleware allways shows isAuthorized

func AlreadyLoggedIn(r *http.Request) bool {
	c, err := r.Cookie(COOKIE_NAME)
	if err != nil {
		return false
	}
	sess, _ := sessionStore.data[c.Value]
	if sess.Username != "" {
		return true
	}
	return false
}

func verifyContent(s string) bool {
	letters := 0

	var symbol, sevenOrTen bool
	for _, c := range s {

		if unicode.IsLetter(c) {
			symbol = true
			letters++
		}

	}
	sevenOrTen = letters >= 8
	if sevenOrTen && symbol {
		return true
	}
	return false
}
