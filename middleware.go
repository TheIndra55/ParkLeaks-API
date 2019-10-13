package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

// MiddleWare is general middleware which adds right headers
func MiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		next.ServeHTTP(w, r)
	})
}

// PostMiddleware is used to return before any additional paths if an post doesn't exist
func PostMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		docs, err := Db.Query("SELECT COUNT(*) FROM `posts` WHERE `id` = ? AND `public` = 1", mux.Vars(r)["post"])
		defer docs.Close()
		if err != nil {
			w.WriteHeader(500)
			WriteErrors(500, []string{"An internal server error occured", "Something went wrong while retrieving the data"}, w)
			return
		}

		var count int
		docs.Next()
		docs.Scan(&count)

		if count > 0 {
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(404)
			WriteError(404, "The requested post was not found", w)
		}
	})
}

// NotFound returns JSON based response when a resource isn't found
func NotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	WriteError(404, "The requested resource was not found", w)
}

// MethodNotAllowed returns JSON based response on MethodNotAllowed responses
func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	WriteError(405, "That method is not allowed on this resource", w)
}
