package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
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
	w.Write([]byte("signing up"))

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
