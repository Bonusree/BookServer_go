package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/goccy/go-json"
)

// JWT secret and token instance
var tokenAuth = jwtauth.New("HS256", []byte("secret_key"), nil)

// Data Structures
type Author struct {
	Name     string `json:"name"`
	Home     string `json:"home"`
	Age      string `json:"age"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Book struct {
	Name    string   `json:"book_name,omitempty"`
	Authors []Author `json:"authors"`
	ISBN    string   `json:"isbn,omitempty"`
	Genre   string   `json:"genre,omitempty"`
	Pub     string   `json:"pub,omitempty"`
}

type AuthorBooks struct {
	Author Author `json:"author"`
	Books  []Book `json:"books"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type BookDB map[string]Book
type AuthorDB map[string]AuthorBooks
type CredDB map[string]string

var BookList BookDB
var AuthorList AuthorDB
var CredList CredDB

// Helper functions
func CapToSmall(s string) string { return strings.ToLower(s) }
func RmSpaces(s string) string   { return strings.ReplaceAll(s, " ", "") }
func SmStr(s string) string      { return CapToSmall(RmSpaces(s)) }

// Initialize with admin
func Init() {
	BookList = make(BookDB)
	AuthorList = make(AuthorDB)
	CredList = make(CredDB)

	admin := Author{Name: "Admin User",
		Home:     "HQ",
		Age:      "50",
		Username: "admin",
		Password: "admin123"}
	key := SmStr(admin.Username)
	AuthorList[key] = AuthorBooks{Author: admin, Books: []Book{}}
	CredList[admin.Username] = admin.Password
}

// Handlers
func SignUp(w http.ResponseWriter, r *http.Request) {
	var req Author
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Home == "" || req.Age == "" || req.Username == "" || req.Password == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}
	if _, exists := CredList[req.Username]; exists {
		http.Error(w, "Username already exists", http.StatusBadRequest)
		return
	}

	key := SmStr(req.Username) // key by username
	AuthorList[key] = AuthorBooks{Author: req, Books: []Book{}}
	CredList[req.Username] = req.Password

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Author registered successfully"))
}

func Login(w http.ResponseWriter, r *http.Request) {
	var cred Credentials
	if err := json.NewDecoder(r.Body).Decode(&cred); err != nil {
		http.Error(w, "Cannot decode data", http.StatusBadRequest)
		return
	}
	password, exists := CredList[cred.Username]
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

func ListAuthors(w http.ResponseWriter, r *http.Request) {
	var authors []Author
	for _, ab := range AuthorList {
		authors = append(authors, ab.Author)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authors)
}

func AddBooks(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil || claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	authorUsername, ok := claims["username"].(string)
	if !ok || authorUsername == "" {
		http.Error(w, "Invalid token data", http.StatusBadRequest)
		return
	}

	var books []Book
	if err := json.NewDecoder(r.Body).Decode(&books); err != nil {
		http.Error(w, "Invalid book data", http.StatusBadRequest)
		return
	}

	authorKey := SmStr(authorUsername)
	authorBooks, exists := AuthorList[authorKey]
	if !exists {
		http.Error(w, "Author not found", http.StatusNotFound)
		return
	}

	for _, book := range books {
		book.Authors = append(book.Authors, authorBooks.Author)
		BookList[book.ISBN] = book
		authorBooks.Books = append(authorBooks.Books, book)
	}
	AuthorList[authorKey] = authorBooks

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Books added successfully"))
}

func GetBooks(w http.ResponseWriter, r *http.Request) {
	var books []Book
	for _, b := range BookList {
		books = append(books, b)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func DeleteAuthor(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	key := SmStr(name)
	if _, exists := AuthorList[key]; !exists {
		http.Error(w, "Author not found", http.StatusNotFound)
		return
	}
	delete(AuthorList, key)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Author deleted successfully"))
}

func main() {
	Init()
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Logger, middleware.Recoverer, middleware.URLFormat)

	// Public routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Welcome to the bookstore API!")
	})
	r.Post("/signup", SignUp)
	r.Post("/login", Login)
	r.Get("/logout", Logout)
	r.Get("/authors", ListAuthors)
	r.Get("/books", GetBooks)
	// Protected routes
	r.Group(func(protected chi.Router) {
		protected.Use(jwtauth.Verifier(tokenAuth))
		protected.Use(jwtauth.Authenticator(tokenAuth))
		protected.Post("/addbooks", AddBooks)
		protected.Delete("/author/{name}", DeleteAuthor)
		//protected.Get("/books", GetBooks)
	})

	log.Println("Server running at http://localhost:8080")
	http.ListenAndServe("localhost:8080", r)
}
