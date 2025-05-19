package apiHandler

import (
	"fmt"
	"github.com/Bonusree/BookServer_go/authHandler"
	"github.com/Bonusree/BookServer_go/dataHandler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/goccy/go-json"
	"log"
	"net/http"
)

func ListAuthors(w http.ResponseWriter, r *http.Request) {
	var authors []dataHandler.Author
	for _, ab := range dataHandler.AuthorList {
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

	var books []dataHandler.Book
	if err := json.NewDecoder(r.Body).Decode(&books); err != nil {
		http.Error(w, "Invalid book data", http.StatusBadRequest)
		return
	}

	authorKey := dataHandler.SmStr(authorUsername)
	authorBooks, exists := dataHandler.AuthorList[authorKey]
	if !exists {
		http.Error(w, "Author not found", http.StatusNotFound)
		return
	}

	for _, book := range books {
		book.Authors = append(book.Authors, authorBooks.Author)
		dataHandler.BookList[book.ISBN] = book
		authorBooks.Books = append(authorBooks.Books, book)
	}
	dataHandler.AuthorList[authorKey] = authorBooks

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Books added successfully"))
}

func GetBooks(w http.ResponseWriter, r *http.Request) {
	var books []dataHandler.Book
	for _, b := range dataHandler.BookList {
		books = append(books, b)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}
func UpdateBook(w http.ResponseWriter, r *http.Request) {

	_, claims, _ := jwtauth.FromContext(r.Context())
	username, ok := claims["username"].(string)
	if !ok || username == "" {
		http.Error(w, "Invalid token data", http.StatusBadRequest)
		return
	}

	isbn := chi.URLParam(r, "isbn")
	if isbn == "" {
		http.Error(w, "ISBN is required", http.StatusBadRequest)
		return
	}

	book, exists := dataHandler.BookList[isbn]
	if !exists {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	var updateData struct {
		Name  *string `json:"book_name,omitempty"`
		Genre *string `json:"genre,omitempty"`
		Pub   *string `json:"pub,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	if updateData.Name != nil {
		book.Name = *updateData.Name
	}
	if updateData.Genre != nil {
		book.Genre = *updateData.Genre
	}
	if updateData.Pub != nil {
		book.Pub = *updateData.Pub
	}

	dataHandler.BookList[isbn] = book

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Book updated successfully"))
}
func DeleteBook(w http.ResponseWriter, r *http.Request) {
	isbn := chi.URLParam(r, "isbn")

	_, exists := dataHandler.BookList[isbn]
	if !exists {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	_, claims, _ := jwtauth.FromContext(r.Context())
	username, ok := claims["username"].(string)
	if !ok || username == "" {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	authorKey := dataHandler.SmStr(username)
	authorData, found := dataHandler.AuthorList[authorKey]
	if !found {
		http.Error(w, "Author not found", http.StatusNotFound)
		return
	}

	newBookList := []dataHandler.Book{}
	for _, b := range authorData.Books {
		if b.ISBN != isbn {
			newBookList = append(newBookList, b)
		}
	}
	authorData.Books = newBookList
	dataHandler.AuthorList[authorKey] = authorData

	// Delete the book from BookList
	delete(dataHandler.BookList, isbn)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Book deleted successfully"))
}

func RunServer(Port int) {
	dataHandler.Init()
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Logger, middleware.Recoverer, middleware.URLFormat)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Welcome to the bookstore API!")
	})
	r.Post("/signup", authHandler.SignUp)
	r.Post("/login", authHandler.Login)
	r.Get("/logout", authHandler.Logout)
	r.Get("/authors", ListAuthors)
	r.Get("/books", GetBooks)
	r.Group(func(protected chi.Router) {
		protected.Use(jwtauth.Verifier(authHandler.TokenAuth))
		protected.Use(jwtauth.Authenticator(authHandler.TokenAuth))
		protected.Post("/addbooks", AddBooks)
		protected.Put("/updatebook/{isbn}", UpdateBook)
		protected.Delete("/deletebook/{isbn}", DeleteBook)
	})

	log.Println("Server running at http://localhost:8080")
	http.ListenAndServe("localhost:8080", r)
}
