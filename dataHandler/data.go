package dataHandler

import (
	"strings"
)

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

func CapToSmall(s string) string { return strings.ToLower(s) }
func RmSpaces(s string) string   { return strings.ReplaceAll(s, " ", "") }
func SmStr(s string) string      { return CapToSmall(RmSpaces(s)) }

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
