package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"log"
	"net/http"
	"strings"
	"time"
)

// JWT secret and token instance
var tokenAuth = jwtauth.New("HS256", []byte("secret_key"), nil)

type Author struct {
	Name string `json:"name,omitempty"`
	Home string `json:"home"`
	Age  string `json:"age"`
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

func CapToSmall(s string) string {
	return strings.ToLower(s)
}
func RmSpaces(s string) string {
	return strings.ReplaceAll(s, " ", "")
}
func SmStr(s string) string {
	return CapToSmall(RmSpaces(s))
}

// Initialize mock data
func Init() {
	author1 := Author{Name: "Temp Author 1", Home: "America", Age: "45"}
	author2 := Author{Name: "Temp Author 2", Home: "Bangladesh", Age: "45"}

	data1 := Book{
		Name:    "Temp Book 1",
		Authors: []Author{author1, author2},
		ISBN:    "ISBN1",
		Genre:   "Fiction",
		Pub:     "Demo",
	}
	data2 := Book{
		Name:    "Temp Book 2",
		Authors: []Author{author1},
		ISBN:    "ISBN2",
		Genre:   "Fiction",
		Pub:     "Demo",
	}

	User := Credentials{Username: "user", Password: "pass"}

	BookList = make(BookDB)
	AuthorList = make(AuthorDB)
	CredList = make(CredDB)

	BookList[data1.ISBN] = data1
	BookList[data2.ISBN] = data2

	var ab1 AuthorBooks
	ab1.Author = author1
	ab1.Books = append(ab1.Books, data2)

	var ab2 AuthorBooks
	ab2.Author = author2
	ab2.Books = append(ab2.Books, data1)

	AuthorList[SmStr(author1.Name)] = ab1
	AuthorList[SmStr(author2.Name)] = ab2

	CredList[User.Username] = User.Password
}

// GET /books - show all books
func GetBooks(w http.ResponseWriter, _ *http.Request) {
	//log.Println("GET /books was hit")
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(BookList)
	fmt.Fprintln(w, err)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//w.WriteHeader(http.StatusOK)
}
func BookGeneralized(w http.ResponseWriter, _ *http.Request) {
	var readString []string
	for _, str := range BookList {
		readString = append(readString, str.Name)
	}
	fmt.Fprintln(w, readString)
	resp := strings.Join(readString, "\n")
	if resp == "" {
		http.Error(w, "No Books found", http.StatusNotFound)
		return
	}
	_, err := w.Write([]byte(resp))

	if err != nil {
		http.Error(w, "Cannot Write Response", http.StatusInternalServerError)
		return
	}
}
func NewBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(&book)
	if err != nil {
		http.Error(w, "Cannot Decode the data", http.StatusBadRequest)
		return
	}
	if len(book.Name) == 0 || len(book.ISBN) == 0 || len(book.Authors) == 0 {
		http.Error(w, "Invalid Data Entry", http.StatusBadRequest)
		return
	}
	flag := false
	for _, data := range book.Authors {
		if len(data.Name) == 0 {
			flag = true
			break
		}
	}
	if flag == true {
		http.Error(w, "Author name can't be empty", http.StatusBadRequest)
		return
	}

	//for _, author := range book.Authors {
	//	name := author.Name
	//	_, ok := AuthorList[SmStr(name)]
	//	var ab AuthorBooks
	//	if ok == false {
	//		ab.Author = author
	//		ab.Books = append(ab.Books, book.ISBN)
	//		AuthorList[SmStr(name)] = ab
	//		continue
	//	}
	//	ab = AuthorList[SmStr(name)]
	//	ab.Books = append(ab.Books, book.ISBN)
	//}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Data added successfully"))
	if err != nil {
		http.Error(w, "Can not Write Data", http.StatusInternalServerError)
		return
	}

	BookList[book.ISBN] = book
	//GetBooks(w, r)
}

// POST /login - authenticate and set JWT
func Login(w http.ResponseWriter, r *http.Request) {
	var cred Credentials
	err := json.NewDecoder(r.Body).Decode(&cred)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	password, ok := CredList[cred.Username]
	if !ok || password != cred.Password {
		http.Error(w, "Invalid username or password", http.StatusBadRequest)
		return
	}

	expTime := time.Now().Add(15 * time.Minute)
	_, tokenString, err := tokenAuth.Encode(map[string]interface{}{
		"aud": "Bonusree",
		"exp": expTime.Unix(),
	})

	if err != nil {
		http.Error(w, "Cannot generate JWT", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "jwt",
		Value:   tokenString,
		Expires: expTime,
	})
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Login successful")
}
func Logout(w http.ResponseWriter, _ *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "jwt",
		Expires: time.Now(), // Expire the cookie immediately
	})
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Logout successful")
}

func main() {
	Init()

	r := chi.NewRouter()
	//CredList = CredDB{
	//	"user1": "password",
	//}

	// Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)

	// Routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to the home page!")
	})
	r.Get("/getbook", GetBooks)
	r.Get("/bookgeneralized", BookGeneralized)
	r.Get("/newbook", NewBook)
	r.Post("/login", Login)
	r.Post("/logout", Logout)

	log.Println("Server running at http://localhost:8080")
	http.ListenAndServe("localhost:8080", r)
}
