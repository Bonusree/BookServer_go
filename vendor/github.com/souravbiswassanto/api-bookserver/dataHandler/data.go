package dataHandler

import "strings"

// Author struct holds common information of an Author
type Author struct {
	Name string `json:"name,omitempty"`
	Home string `json:"home"`
	Age  string `json:"age"`
}

// AuthorBooks Use Composition to store Books
// ISBN which can be different for each
type AuthorBooks struct {
	Author `json:"author"`
	Books  []string `json:"books"`
}

// Book store Book information and the Authors who authored it
type Book struct {
	Name    string   `json:"book_name,omitempty"`
	Authors []Author `json:"authors"`
	ISBN    string   `json:"isbn,omitempty"`
	Genre   string   `json:"genre"`
	Pub     string   `json:"publisher"`
}

// Credentials Stores Login Credentials
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// BookDB AuthorDB CredDB are databases
type BookDB map[string]Book
type AuthorDB map[string]AuthorBooks
type CredDB map[string]string

// BookList AuthorList CredList are DB Instances
var BookList BookDB

// TODO : var MyList []Book

var AuthorList AuthorDB
var CredList CredDB

func Init() {
	author1 := Author{
		Name: "temp author 1",
		Home: "America",
		Age:  "45",
	}
	author2 := Author{
		Name: "temp author 2",
		Home: "Bangladesh",
		Age:  "45",
	}

	data1 := Book{
		Name: "temp book 1",
		Authors: []Author{
			author1,
			author2,
		},
		ISBN:  "ISBN1",
		Genre: "Fiction",
		Pub:   "Demo",
	}

	data2 := Book{
		Name: "temp book 2",
		Authors: []Author{
			author1,
		},
		ISBN:  "ISBN2",
		Genre: "Fiction",
		Pub:   "Demo",
	}

	User := Credentials{
		Username: "user",
		Password: "pass",
	}

	BookList = make(BookDB)
	AuthorList = make(AuthorDB)
	CredList = make(CredDB)

	var ab1 AuthorBooks
	ab1.Author = author1
	ab1.Books = append(ab1.Books, data1.ISBN)
	ab1.Books = append(ab1.Books, data2.ISBN)

	var ab2 AuthorBooks
	ab2.Author = author2
	ab2.Books = append(ab2.Books, data1.ISBN)

	AuthorList[SmStr(author1.Name)] = ab1
	AuthorList[SmStr(author2.Name)] = ab2

	BookList[data1.ISBN] = data1
	BookList[data2.ISBN] = data2

	CredList[User.Username] = User.Password

	return
}

// CapToSmall converts string from Capital to Small
// RmSpace Removes Spaces from string
// SmStr uses above two function to remake a string
// Which removes spaces and also apply CapToSmall
func capToSmall(s string) string {
	return strings.ToLower(s)
}
func rmSpace(s string) string {
	return strings.ReplaceAll(s, " ", "")
}
func SmStr(s string) string {
	return capToSmall(rmSpace(s))
}
