package authHandler

import (
	"github.com/Bonusree/BookServer_go/dataHandler"
	"github.com/go-chi/jwtauth/v5"
	"github.com/goccy/go-json"
	"net/http"
	"time"
)

var TokenAuth = jwtauth.New("HS256", []byte("secret_key"), nil)

func SignUp(w http.ResponseWriter, r *http.Request) {
	var req dataHandler.Author
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Home == "" || req.Age == "" || req.Username == "" || req.Password == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}
	if _, exists := dataHandler.CredList[req.Username]; exists {
		http.Error(w, "Username already exists", http.StatusBadRequest)
		return
	}

	key := dataHandler.SmStr(req.Username)
	dataHandler.AuthorList[key] = dataHandler.AuthorBooks{Author: req, Books: []dataHandler.Book{}}
	dataHandler.CredList[req.Username] = req.Password

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Author registered successfully"))
}

func Login(w http.ResponseWriter, r *http.Request) {
	var cred dataHandler.Credentials
	if err := json.NewDecoder(r.Body).Decode(&cred); err != nil {
		http.Error(w, "Cannot decode data", http.StatusBadRequest)
		return
	}
	password, exists := dataHandler.CredList[cred.Username]
	if !exists || password != cred.Password {
		http.Error(w, "Invalid credentials", http.StatusBadRequest)
		return
	}
	expirationTime := time.Now().Add(15 * time.Minute)
	_, tokenString, err := tokenAuth.Encode(map[string]interface{}{
		"username": cred.Username,
		"exp":      expirationTime.Unix(),
	})
	if err != nil {
		http.Error(w, "Could not generate JWT", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "jwt",
		Value:   tokenString,
		Expires: expirationTime,
		Path:    "/",
	})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Login successful"))
}

func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logout successful"))
}
