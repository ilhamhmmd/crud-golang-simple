package handler

import (
	"crud-golang-simple/model"
	"database/sql"	
	"fmt"
	"html/template"
	"net/http"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/go-sql-driver/mysql"	
	"github.com/kataras/go-sessions"
	// "os"
)

type user struct {
	ID int
	Username string
	FirstName string
	LastName string
	Password string
}

var username, password, host, nameDB, defaultDB string
var db *sql.DB
var err error

func init() {
	username = "root"
	password = ""
	host = "localhost"
	nameDB = "golang_crud"
	defaultDB = "mysql"
}

func QueryUser(username string) user {
	var users = user{}
	err = db.QueryRow(
		`SELECT id,
		username,
		first_name,
		last_name,
		password 
		FROM users WHERE username=?`, username).
		
		Scan(
			&users.ID,
			&users.Username,
			&users.FirstName,
			&users.LastName,
			&users.Password,
		)		
		return users
}

func checkErr(w http.ResponseWriter, r *http.Request, err error) bool {
	if err != nil {
		fmt.Println(r.Host + r.URL.Path)

		http.Redirect(w, r, r.Host + r.URL.Path, 301)
		return false
	}
	return true
}

func Home(w http.ResponseWriter, r *http.Request) {
	db, err = model.ConnectDB(username, password, host, nameDB)

	if err != nil {
		return
	}

	defer db.Close()
	
	session := sessions.Start(w, r)
	if len(session.GetString("username")) == 0 {
		http.Redirect(w, r, "/login", 301)
	}

	var data = map[string] string {
		"username": session.GetString("username"),
		"message": "Welcome to the Go !",
	}

	var t, err = template.ParseFiles("views/home.html")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	t.Execute(w, data)
	return
}

func Register(w http.ResponseWriter, r *http.Request) {
	db, err = model.ConnectDB(username, password, host, nameDB)

	if err != nil {
		return
	}

	defer db.Close()

	if r.Method != "POST" {
		http.ServeFile(w,r, "views/register.html")
		return
	}

	username := r.FormValue("email")
	first_name := r.FormValue("first_name")	
	last_name := r.FormValue("last_name")	
	password := r.FormValue("password")

	users := QueryUser(username)

	if (user{}) == users {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		if len(hashedPassword) != 0 && checkErr(w, r, err) {
			stmt, err := db.Prepare("INSERT INTO users SET username=?, password=?, first_name=?, last_name=?")
			if err == nil {
				_, err := stmt.Exec(&username, &hashedPassword, &first_name, &last_name)

				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
		}
	} else {
		http.Redirect(w, r, "/register", 302)
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	db, err = model.ConnectDB(username, password, host, nameDB)

	if err != nil {
		return
	}

	defer db.Close()

	session := sessions.Start(w, r)
	if len(session.GetString("username")) != 0 && checkErr(w, r, err) {
		http.Redirect(w, r, "/", 302)
	}

	if r.Method != "POST" {
		http.ServeFile(w, r, "views/login.html")
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	users := QueryUser(username)

	//deskripsi dan compare password
	var password_test = bcrypt.CompareHashAndPassword([]byte(users.Password), []byte(password))
	if password_test == nil {
		//login sukses
		session := sessions.Start(w, r)
		session.Set("username", users.Username)
		session.Set("name", users.FirstName)
		http.Redirect(w, r, "/", 302)
		} else {
			//Login Gagal
			http.Redirect(w, r, "/login", 302)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	session.Clear()
	sessions.Destroy(w, r)
	http.Redirect(w, r, "/", 302)
}