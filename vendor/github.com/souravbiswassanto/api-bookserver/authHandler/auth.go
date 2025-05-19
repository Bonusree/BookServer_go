package authHandler

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/souravbiswassanto/api-bookserver/dataHandler"
	"net/http"
	"time"
)

// Secret is the secret key used for java web token verify signature
// see this video for better understanding
// https://www.youtube.com/watch?v=7Q17ubqLfaM&t=238s
var Secret = []byte("this_is_a_secret_key")
var TokenAuth *jwtauth.JWTAuth

// InitToken initiates the algorithm that we will use for jwt
// It also initiates the secret token
func InitToken() {
	TokenAuth = jwtauth.New(string(jwa.HS256), Secret, nil)
}

func Login(w http.ResponseWriter, r *http.Request) {
	var cred dataHandler.Credentials
	err := json.NewDecoder(r.Body).Decode(&cred)
	fmt.Println(cred)
	if err != nil {
		http.Error(w, "Can not Decode the data", http.StatusBadRequest)
		return
	}

	password, okay := dataHandler.CredList[cred.Username]

	if okay == false {
		http.Error(w, "Username do not exist", http.StatusBadRequest)
		return
	}

	if password != cred.Password {
		http.Error(w, "Password not matching", http.StatusBadRequest)
		return
	}
	et := time.Now().Add(15 * time.Minute)
	_, tokenString, err := TokenAuth.Encode(map[string]interface{}{
		"aud": "Saurov Biswas",
		"exp": et.Unix(),
		// Here few more registered field and also self-driven field can be added
	})
	fmt.Println(tokenString)
	if err != nil {
		http.Error(w, "Can not generate jwt", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "jwt",
		Value:   tokenString,
		Expires: et,
	})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Login Successful"))
}

func Logout(w http.ResponseWriter, _ *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "jwt",
		Expires: time.Now(),
	})
	w.WriteHeader(http.StatusOK)
}
