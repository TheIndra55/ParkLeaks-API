package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

// Response is top level root of api response
type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
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
	User User   `json:"user"`
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
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Vip   bool   `json:"vip"`
	Staff bool   `json:"staff"`
}

func main() {
	router := mux.NewRouter()

	OpenDatabase()

	postsRouter := router.PathPrefix("/posts/").Subrouter()
	postsRouter.HandleFunc("/{post}", HandlePost)
	postsRouter.HandleFunc("/{post}/comments", HandleComments).Methods("GET")
	postsRouter.HandleFunc("/{post}/comments", HandleComment).Methods("POST")
	postsRouter.HandleFunc("/{post}/vote", HandleVote)
	postsRouter.Use(PostMiddleware)

	usersRouter := mux.NewRouter()
	//usersRouter.HandleFunc("/")

	router.HandleFunc("/posts", HandlePosts)
	router.Handle("/", postsRouter)
	router.Handle("/users/", usersRouter)

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
