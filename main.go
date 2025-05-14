package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/goccy/go-json"
)

// JWT secret and token instance
var tokenAuth = jwtauth.New("HS256", []byte("secret_key"), nil)

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

// We store authors by key = processed name
// Value = AuthorBooks (including books by that author)
type AuthorDB map[string]AuthorBooks

// Credentials for login/signup
type CredDB map[string]string

var BookList BookDB
var AuthorList AuthorDB
var CredList CredDB

func CapToSmall(s string) string {
	return strings.ToLower(s)
}

func RmSpaces(s string) string {
	return strings.ReplaceAll(s, " ", "")
}

func SmStr(s string) string {
	return CapToSmall(RmSpaces(s))
}

func Init() {
	BookList = make(BookDB)
	AuthorList = make(AuthorDB)
	CredList = make(CredDB)

	admin := Author{
		Name:     "Admin User",
		Home:     "HQ",
		Age:      "50",
		Username: "admin",
		Password: "admin123",
	}

	AuthorList[SmStr(admin.Name)] = AuthorBooks{
		Author: admin,
		Books:  []Book{},
	}
	CredList[admin.Username] = admin.Password
}

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

	author := Author{Name: req.Name, Home: req.Home, Age: req.Age, Username: req.Username, Password: req.Password}
	key := SmStr(author.Name)
	AuthorList[key] = AuthorBooks{Author: author, Books: []Book{}}
	CredList[author.Username] = author.Password

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Author registered successfully"))
}
func Login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	pass, exists := CredList[creds.Username]
	if !exists || pass != creds.Password {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"username": creds.Username})
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"token": "%s"}`, tokenString)))
}

func Logout(w http.ResponseWriter, r *http.Request) {
	// Stateless JWT: instruct client to discard token
	w.Write([]byte("Logged out. Please discard the token on client side."))
}
func ListAuthors(w http.ResponseWriter, r *http.Request) {
	var authors []Author
	for _, ab := range AuthorList {
		authors = append(authors, ab.Author)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authors)
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
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to the bookstore API!")
	})
	r.Post("/signup", SignUp)
	r.Post("/login", Login)
	r.Get("/logout", Logout)
	r.Get("/authors", ListAuthors)
	r.Delete("/author/{name}", DeleteAuthor)

	log.Println("Server running at http://localhost:8080")
	http.ListenAndServe("localhost:8080", r)
}
