package main

import (
	"github.com/gorilla/mux"
)

// Response is top level root of api response
type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

func main() {
	router := mux.NewRouter()

	postsRouter := mux.NewRouter()
	postsRouter.HandleFunc("/{post}", HandlePost)
	postsRouter.HandleFunc("/{post}/comments", HandleComments)
	postsRouter.HandleFunc("/{post}/vote", HandleVote)

	usersRouter := mux.NewRouter()
	//usersRouter.HandleFunc("/")

	router.HandleFunc("/posts", HandlePosts)
	router.Handle("/posts/", postsRouter)
	router.Handle("/users/", usersRouter)

	OpenDatabase()
}
