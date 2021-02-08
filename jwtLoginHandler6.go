package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	jwt "github.com/dgrijalva/jwt-go"
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

func generateToken(user User) (string, error) {
	var err error
	secret := "secret"
	//The member names within the JWT Claims Set are referred to as Claim Names.
	// The corresponding values are referred to as Claim Values.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": user.Email,
		"iss":   "course",
	})
	// a JWT is header.payload.secret

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Println(err)
		log.Fatalln(err)
	}
	return tokenString, nil
}

var db *sql.DB

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

	stmt := "insert into users (email, password) values ($1, $2) RETURNING id;"
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
	var user User
	var jwt JWT
	var error Error
	json.NewDecoder(r.Body).Decode(&user)
	spew.Dump(user)
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

	hashedPasswordFromUser := user.Password

	row := db.QueryRow("select * from users1 where email=$1", user.Email)
	err := row.Scan(&user.ID, &user.Email, &user.Password)

	if err != nil {
		if err == sql.ErrNoRows {
			error.Message = "The user does not exist"
			respondWithError(w, http.StatusBadRequest, error)
			return
		} else {
			log.Fatal(err)
		}
	}
	spew.Dump(user)

	hashedPasswordFromDB := user.Password

	err = bcrypt.CompareHashAndPassword([]byte(hashedPasswordFromDB), []byte(hashedPasswordFromUser))
	fmt.Println(err)

	if err != nil {
		error.Message = "Invalid Password"
		respondWithError(w, http.StatusUnauthorized, error)
		return
	}

	token, err := generateToken(user)

	if err != nil {
		log.Fatal(err)
	}
	w.WriteHeader(http.StatusOK)
	jwt.Token = token
	responseJSON(w, jwt)

}

func protectedEndpoint(w http.ResponseWriter, r *http.Request) {
	fmt.Println("protectedEndpoint invoked")
}

func TokenVerifyMiddleWare(next http.HandlerFunc) http.HandlerFunc {
	fmt.Println("TokenVerifyMiddleWare invoked")
	return nil
}

func main() {
	//fmt.Println("Hi")

	pgUrl, err := pq.ParseURL("postgres://hxfydbwp:Q2hfldxp88X8d6WOW50KnNsaD42zbT0A@ruby.db.elephantsql.com:5432/hxfydbwp")

	if err != nil {
		log.Fatalln(err)
	}

	//fmt.Println(db)

	db, err = sql.Open("postgres", pgUrl)

	if err != nil {
		log.Fatalln(err)
	}

	//fmt.Println(db)

	router := mux.NewRouter()
	router.HandleFunc("/signup", signup).Methods("POST")
	router.HandleFunc("/login", login).Methods("POST")
	router.HandleFunc("/protected", TokenVerifyMiddleWare(protectedEndpoint)).Methods("GET")

	log.Println("listening on port 8000 ....")
	log.Fatal(http.ListenAndServe(":8000", router))
}
