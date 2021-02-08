package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

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

	/*
			.1.1.  "iss" (Issuer) Claim

		   The "iss" (issuer) claim identifies the principal that issued the
		   JWT.  The processing of this claim is generally application specific.
		   The "iss" value is a case-sensitive string containing a StringOrURI
		   value.  Use of this claim is OPTIONAL.
	*/
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

	user.Password = string(hash)

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

	if err != nil {
		error.Message = "Invalid Password"
		respondWithError(w, http.StatusUnauthorized, error)
		return
	}

	token, err := generateToken(user)
	fmt.Println(token)

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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var errorObject Error
		authHeader := r.Header.Get("Authorization")
		bearerToken := strings.Split(authHeader, " ")

		if len(bearerToken) == 2 {
			authToken := bearerToken[1]

			token, error := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("There was an error")
				}

				return []byte("secret"), nil
			})

			if error != nil {
				errorObject.Message = error.Error()
				respondWithError(w, http.StatusUnauthorized, errorObject)
				return
			}

			if token.Valid {
				next.ServeHTTP(w, r)
			} else {
				errorObject.Message = error.Error()
				respondWithError(w, http.StatusUnauthorized, errorObject)
				return
			}
		} else {
			errorObject.Message = "Invalid token."
			respondWithError(w, http.StatusUnauthorized, errorObject)
			return
		}
	})
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
