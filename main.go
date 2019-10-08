package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
)

// Response is top level root of api response
type Response struct {
	Code   int         `json:"code"`
	Data   interface{} `json:"data,omitempty"`
	Errors []string    `json:"errors,omitempty"`
}

// Post represents a post
type Post struct {
	ID       int      `json:"id"`
	Title    string   `json:"title"`
	Text     string   `json:"text"`
	Images   []string `json:"images"`
	Stats    Stats    `json:"stats"`
	Verified bool     `json:"verified"`
	Date     string   `json:"date"`
}

// Comment represents a comment on a post
type Comment struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	User *User  `json:"user,omitempty"`
	Date string `json:"date"`
}

// Stats stores stats about the post like views, comment count and score
type Stats struct {
	Views    int `json:"views"`
	Comments int `json:"comments"`
	Score    int `json:"score"`
}

// User represents an author on post or comment
type User struct {
	ID      *int   `json:"id"`
	Name    string `json:"name"`
	Vip     bool   `json:"vip"`
	Staff   bool   `json:"staff"`
	Address string `json:"-"`
}

func main() {
	router := mux.NewRouter()

	OpenDatabase()

	postsRouter := router.PathPrefix("/posts/").Subrouter()
	postsRouter.HandleFunc("/{post}", HandlePost).Methods("GET")
	postsRouter.HandleFunc("/{post}/comments", HandleComments).Methods("GET")
	postsRouter.HandleFunc("/{post}/comments", HandleComment).Methods("POST")
	postsRouter.HandleFunc("/{post}/vote", HandleVote).Methods("POST")
	postsRouter.Use(PostMiddleware)

	usersRouter := router.PathPrefix("/users/").Subrouter()
	usersRouter.HandleFunc("/{user}", HandleUser).Methods("GET")

	router.HandleFunc("/posts", HandlePosts).Methods("GET")
	router.Handle("/", postsRouter)
	router.Handle("/", usersRouter)

	router.NotFoundHandler = http.HandlerFunc(NotFound)
	router.MethodNotAllowedHandler = http.HandlerFunc(MethodNotAllowed)
	router.Use(MiddleWare)

	http.ListenAndServe(":8080", router)
}

// WriteResponse writes an response using the Response struct
func WriteResponse(code int, data interface{}, w http.ResponseWriter) {
	res := Response{
		Code: code,
		Data: data,
	}

	out, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	io.WriteString(w, string(out))
}

// WriteErrors writes an errored response with multiple errors
func WriteErrors(code int, errors []string, w http.ResponseWriter) {
	res := Response{
		Code:   code,
		Errors: errors,
	}

	out, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	io.WriteString(w, string(out))
}

// WriteError writes and error response with a single error
func WriteError(code int, err string, w http.ResponseWriter) {
	WriteErrors(code, []string{err}, w)
}
