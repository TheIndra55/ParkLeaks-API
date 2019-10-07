package main

import "net/http"

// HandlePosts function serves /posts which returns all posts
func HandlePosts(w http.ResponseWriter, r *http.Request) {

}

// HandlePost serves a single post
func HandlePost(w http.ResponseWriter, r *http.Request) {

}

// HandleComments services comments for a single post
func HandleComments(w http.ResponseWriter, r *http.Request) {

}

// HandleComment takes care of POST on /comments which posts a comment
func HandleComment(w http.ResponseWriter, r *http.Request) {

}

// HandleVote takes care of POST on /vote which upvotes or downvotes a post
func HandleVote(w http.ResponseWriter, r *http.Request) {

}
