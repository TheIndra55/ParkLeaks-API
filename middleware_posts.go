package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

// PostMiddleware is used to return before any additional paths if an post doesn't exist
func PostMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		docs, err := Db.Query("SELECT COUNT(*) FROM `posts` WHERE `id` = ? AND `public` = 1", mux.Vars(r)["post"])
		if err != nil {
			w.WriteHeader(500)
			return
		}

		var count int
		docs.Next()
		docs.Scan(&count)

		if count > 0 {
			next.ServeHTTP(w, r)
		} else {
			WriteResponse(404, nil, w)
		}
	})
}
