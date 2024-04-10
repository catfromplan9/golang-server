package main

import (
	"fmt"
	"net/http"
	"time"
	//"io/ioutil"
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"html/template"
)

/*
 *            WARNING! 
 * This file is UNDER CONSTRUCTION!
 *             
 */


func hash_password(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func check_password(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

/*
	USER ACCOUNTS
*/

type Account struct {
	id int

	username string
	email    string
	hash     string

	token        string
	token_issued int

	created  int
	verified int

	class int
}

var template_register *template.Template
var template_login *template.Template

func accounts_init() error {

	register, err := template.ParseFiles(html_path + "/register.html")
	template_register = template.Must(register, err)

	login, err := template.ParseFiles(html_path + "/login.html")
	template_login = template.Must(login, err)

	http.HandleFunc(url_root+"/login", handle_login)
	http.HandleFunc(url_root+"/logout", handle_logout)
	http.HandleFunc(url_root+"/register", handle_register)

	return nil
}


func account_set_level(username string) error {

	_, err := db.Exec("UPDATE account SET power ")
	if err != nil {
		return err
	}
	return nil

}

func check_username(username string) (string, error) {
	var count int
	row := db.QueryRow("SELECT COUNT(*) FROM account WHERE username=?1", username)
	err := row.Scan(&count)
	if err != nil {
		return "", err
	}
	if count > 0 {
		return "Username occupied", nil
	}
	return "", nil
}
func check_email(email string) (string, error) {
	var count int
	row := db.QueryRow("SELECT COUNT(*) FROM account WHERE email=?1", email)
	err := row.Scan(&count)
	if err != nil {
		return "", err
	}
	if count > 0 {
		return "Email already in use", nil
	}
	return "", nil
}

/*
  TOKEN MANAGMENT
*/

func account_token_new(id int) (string, error) {
	token := random_string(128)
	token_issued := time.Now().Unix()
	_, err := db.Exec("UPDATE account SET token=?1, token_issued=?2 WHERE id=?3", token, token_issued, id)
	return token, err
}

func account_token_get(id int) (string, error) {
	var token sql.NullString
	var token_issued sql.NullInt64

	row := db.QueryRow("SELECT token, token_issued FROM account WHERE id=?1", id)
	err := row.Scan(&token, &token_issued)

	if err != nil {
		fmt.Println("Token query failed")
		return "", err
	}

	// TODO check token age
	if !token.Valid {
		token.String, err = account_token_new(id)
		if err != nil {
			fmt.Println("Token creation failed")
			return "", err
		}
		token.Valid = true
	}

	return token.String, nil
}

func account_login(username string, password string) (string, error) {

	rows, err := db.Query("SELECT id, hash FROM account WHERE username=?1", username)

	if err != nil {
		fmt.Println("Username login query failed")
		return "", err
	}

	for rows.Next() {
		var id int
		var hash string
		err := rows.Scan(&id, &hash)
		if err != nil {
			rows.Close()
			return "", err
		}
		ok := check_password(password, hash)
		if ok {
			rows.Close()
			token, err := account_token_get(id)
			if err != nil {
				return "", err
			}
			return token, nil
		}

	}

	rows.Close()
	return "", nil
}

func account_register(
	username string,
	password string,
	email string,
) (string, error) {

	hash, err := hash_password((password))
	if err != nil {
		return "", err
	}

	status, err := check_username(username)
	if err != nil || status != "" {
		return status, err
	}

	status, err = check_email(email)
	if err != nil || status != "" {
		return status, err
	}

	fmt.Println("Hash:", hash)

	created := time.Now().Unix()

	_, err = db.Exec(("INSERT INTO account (username, email, hash, created, verified, power)" +
		"VALUES              (?1,       ?2,    ?3,   ?4,      ?5,        ?6)"),
		username, email, hash, created, 0, 0)

	if err != nil {
		return "", err
	}

	return "", nil
}

func handle_register(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case http.MethodGet:

		fmt.Println("Register")

		data := struct {
			Title  string
			Action string
		}{
			Title:  site_title,
			Action: url_root + "/register",
		}

		template_register.Execute(w, data)

	case http.MethodPost:
		req.ParseForm()
		username := req.FormValue("username")
		password := req.FormValue("password")
		email := req.FormValue("email")

		status, err := account_register(username, password, email)

		if err != nil {
			fmt.Println(err)
			handler_error(w, req, http.StatusInternalServerError)
			return
		}

		//fmt.Fprintf(w, "%v %s\n", num, err);
		fmt.Fprintf(w, "<!doctype html>")
		fmt.Fprintf(w, "%s<br>", status)
		fmt.Fprintf(w, "%s<br>", username)
		fmt.Fprintf(w, "%s<br>", email)
		fmt.Fprintf(w, "%s<br>", password)

	default:
		handler_error(w, req, http.StatusMethodNotAllowed)
	}
}
func handle_login(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case http.MethodGet:

		fmt.Println("Login")

		data := struct {
			Title  string
			Action string
		}{
			Title:  site_title,
			Action: url_root + "/login",
		}

		template_login.Execute(w, data)

	case http.MethodPost:
		req.ParseForm()
		username := req.FormValue("username")
		password := req.FormValue("password")

		start := time.Now().UnixMilli()
		token, err := account_login(username, password)
		fmt.Println("Took", time.Now().UnixMilli()-start, "ms")

		if err != nil {
			fmt.Println(err)
			handler_error(w, req, http.StatusInternalServerError)
			return
		}

		//fmt.Fprintf(w, "%v %s\n", num, err);
		fmt.Fprintf(w, "<!doctype html>")
		fmt.Fprintf(w, "%s<br>", token)
		fmt.Fprintf(w, "%s<br>", username)
		fmt.Fprintf(w, "%s<br>", password)

	default:
		handler_error(w, req, http.StatusMethodNotAllowed)
	}
}


func handle_logout(w http.ResponseWriter, req *http.Request) {
	c := &http.Cookie{
		Name:    "upload_token",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	}
	http.SetCookie(w, c)
	http.Redirect(w, req, "/upload", http.StatusSeeOther)
}
