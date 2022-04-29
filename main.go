package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type User struct {
	gorm.Model

	UserName string `gorm:"type:varchar(20);unique_index"`
	Password string
	Email    string `gorm:"type:varchar(100);unique_index"`
}

type Profile struct {
	gorm.Model

	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	ProfilePicture string `json:"profile_picture"`
	PhoneNumber    string `json:"phone_number"`
	UserID         uint   `json:"user_id"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type SuccessToken struct {
	AccessToken         string `json:"access_token"`
	RefreshToken        string `json:"refresh_token"`
	AccessTokenExpires  string `json:"access_token_expires"`
	RefreshTokenExpires string `json:"refresh_token_expires"`
}

var db *gorm.DB
var err error
var jwtkey = []byte(os.Getenv("JWT_KEY"))

func main() {

	/*
		loading environmental variables
	*/
	dialect := os.Getenv("DIALECT")
	host := os.Getenv("HOST")
	dbport := os.Getenv("DBPORT")
	user := os.Getenv("USER")
	dbName := os.Getenv("NAME")
	password := "zoom20$$" //os.Getenv("PASSWD")

	/*
		Database connection string
	*/
	dbURI := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", host, dbport, user, dbName, password)

	/*
		opening database connection
	*/
	db, err = gorm.Open(dialect, dbURI)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Successfully connected to database")
	}

	/*
		close connection to database when main function finishes
	*/
	defer db.Close()

	/*
		Migrating migrations to the database if they have not been created
	*/
	db.AutoMigrate(&User{}, &Profile{})

	handleRequests()
}

/* Handle API requests*/
func handleRequests() {

	router := mux.NewRouter()

	/*
		API endpoints
	*/
	router.HandleFunc("/signup", signUp).Methods("POST")
	router.HandleFunc("/login", login).Methods("POST")
	router.Handle("/user/{id}", isAuthorized(getUser)).Methods("GET")
	router.Handle("/profile/{id}", isAuthorized(getProfile)).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", router))
}

/*
	Generate JWT token
*/
func generateToken(principal string, duration time.Duration) (string, int64, error) {
	claims := &Claims{
		principal,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(duration).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtkey)

	if err != nil {
		return "", 0, err
	}

	return tokenString, claims.ExpiresAt, nil
}

/* Handle JWT Requests*/
func isAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenString := r.Header.Get("Authorization")

		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)

		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims,
			func(token *jwt.Token) (interface{}, error) {
				return jwtkey, nil
			})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode("Invalid Token")
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(err)
			json.NewEncoder(w).Encode("Bad Request")
			return
		}

		if !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode("Invalid Token")
			return
		}

		endpoint(w, r)

	})
}

/*
	API controllers
*/
func signUp(w http.ResponseWriter, r *http.Request) {
	var user User
	_ = json.NewDecoder(r.Body).Decode(&user)

	createdUser := db.Create(&user)
	err = createdUser.Error

	if err != nil {

		json.NewEncoder(w).Encode(err)

	} else {

		/*if there is no error create profile*/
		profile := &Profile{
			FirstName:      "",
			LastName:       "",
			ProfilePicture: "",
			PhoneNumber:    "07123456789",
			UserID:         user.ID,
		}

		createProfile := db.Create(profile)
		err = createProfile.Error

		if err != nil {
			json.NewEncoder(w).Encode(err)
		}

		json.NewEncoder(w).Encode(user)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	var user User
	_ = json.NewDecoder(r.Body).Decode(&user)

	var dbUser User
	db.Where("user_name = ?", user.UserName).First(&dbUser)

	if dbUser.UserName == user.UserName && dbUser.Password == user.Password {

		accessTokenExpirationTime := time.Duration(30) * time.Minute
		refreshTokenExpirationTime := time.Duration(30*24) * time.Hour

		accessToken, accessTokenExpiresAt, err := generateToken(user.UserName, accessTokenExpirationTime)
		if err != nil {
			json.NewEncoder(w).Encode(http.StatusInternalServerError)
		}

		refreshToken, refreshTokenExpiresAt, err := generateToken(user.UserName, refreshTokenExpirationTime)
		if err != nil {
			json.NewEncoder(w).Encode(http.StatusInternalServerError)
		}

		successtoken := &SuccessToken{
			AccessToken:         accessToken,
			RefreshToken:        refreshToken,
			AccessTokenExpires:  time.Unix(accessTokenExpiresAt, 0).Format(time.RFC3339),
			RefreshTokenExpires: time.Unix(refreshTokenExpiresAt, 0).Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(successtoken)

	} else {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode("Invalid Username or Password")
	}
}

func getUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])

	var user User
	db.First(&user, id)

	if user.ID == 0 {
		json.NewEncoder(w).Encode(user)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func getProfile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])

	var profile Profile
	db.First(&profile, id)

	if profile.ID == 0 {
		json.NewEncoder(w).Encode(profile)
		return
	}

	json.NewEncoder(w).Encode(profile)
}
