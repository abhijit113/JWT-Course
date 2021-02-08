package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type JWT struct {
	Token string `json:"token"`
}

type Error struct {
	Message string `json:"message"`
}

func respondWithError(w http.ResponseWriter, status int, error Error) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(error)
}

func responseJSON(w http.ResponseWriter, data interface{}) {
	json.NewEncoder(w).Encode(data)
}

var db *sql.DB

func main() {
	fmt.Println("Hi")

	pgUrl, err := pq.ParseURL("postgres://hxfydbwp:Q2hfldxp88X8d6WOW50KnNsaD42zbT0A@ruby.db.elephantsql.com:5432/hxfydbwp")

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(db)

	db, err = sql.Open("postgres", pgUrl)

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(db)

	router := mux.NewRouter()
	router.HandleFunc("/signup", signup).Methods("POST")
	router.HandleFunc("/login", login).Methods("POST")
	router.HandleFunc("/protected", TokenVerifyMiddleWare(protectedEndpoint)).Methods("GET")

	log.Println("listening on port 8000 ....")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func signup(w http.ResponseWriter, r *http.Request) {

	var user User
	var error Error

	json.NewDecoder(r.Body).Decode(&user)
	if user.Email == "" {
		//error response needs to be sent
		//http.StatusBadRequest as 400
		error.Message = "Email is missing"
		respondWithError(w, http.StatusBadRequest, error)
		return

	}

	if user.Password == "" {
		//error response needs to be sent
		//http.StatusBadRequest as 400
		error.Message = "password is missing"
		respondWithError(w, http.StatusBadRequest, error)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("user password", user.Password)
	fmt.Println("hashed password", hash)
	user.Password = string(hash)
	fmt.Println("after hashing, the user password", user.Password)
	fmt.Println(user)

	fmt.Println("**********")

	spew.Dump(user)

	stmt := "insert into users1 (email, password) values ($1, $2) RETURNING id;"
	err = db.QueryRow(stmt, user.Email, user.Password).Scan(&user.ID)

	if err != nil {
		error.Message = "DB Server error"
		respondWithError(w, http.StatusInternalServerError, error)
		return
	}

	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	responseJSON(w, user)

}

func login(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("login is succesful"))
}

func protectedEndpoint(w http.ResponseWriter, r *http.Request) {
	fmt.Println("protectedEndpoint invoked")
}

func TokenVerifyMiddleWare(next http.HandlerFunc) http.HandlerFunc {
	fmt.Println("TokenVerifyMiddleWare invoked")
	return nil
}
