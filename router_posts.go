package main

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// HandlePosts function serves /posts which returns all posts
func HandlePosts(w http.ResponseWriter, r *http.Request) {
	rows, err := Db.Query("SELECT `id`, `titel`, `text`, `images`, `verified`, `date` FROM `posts` WHERE `public` = 1 ORDER BY `id` DESC")
	if err != nil {
		w.WriteHeader(500)
		WriteResponse(500, err, w)
		return
	}

	var posts []Post
	for rows.Next() {
		var (
			id                int
			title, text, date string
			images            string
			verified          bool
		)

		rows.Scan(&id, &title, &text, &images, &verified, &date)
		posts = append(posts, Post{
			ID:       id,
			Title:    title,
			Text:     text,
			Verified: verified,
			Date:     date,
			Images:   Split(images, ","),
		})
	}

	WriteResponse(200, posts, w)
}

// Split is a workaround for Go's strings.Split() function which returns [""] instead of [] if no elements
func Split(str string, sep string) []string {
	if str == "" {
		return []string{}
	}

	return strings.Split(str, sep)
}

// HandlePost serves a single post
func HandlePost(w http.ResponseWriter, r *http.Request) {
	rows, err := Db.Query("SELECT `id`, `titel`, `text`, `images`, `verified`, `date` FROM `posts` WHERE `id` = ?", mux.Vars(r)["post"])
	if err != nil {
		w.WriteHeader(500)
		return
	}

	var (
		id                int
		title, text, date string
		images            string
		verified          bool
	)

	rows.Next()
	rows.Scan(&id, &title, &text, &images, &verified, &date)

	WriteResponse(200, Post{
		ID:       id,
		Title:    title,
		Text:     text,
		Verified: verified,
		Date:     date,
		Images:   Split(images, ","),
	}, w)
}

// HandleComments services comments for a single post
func HandleComments(w http.ResponseWriter, r *http.Request) {
	rows, err := Db.Query("SELECT reacties.id, reacties.text, reacties.date, namen.name, namen.team, namen.vip, namen.id as `userid` FROM `reacties` "+
		"INNER JOIN `namen` ON reacties.address=namen.address WHERE `postid` = ? ORDER BY `id` DESC", mux.Vars(r)["post"])

	if err != nil {
		w.WriteHeader(500)
		return
	}

	var comments []Comment
	for rows.Next() {
		var (
			id, userid       int
			text, date, name string
			team, vip        bool
		)

		rows.Scan(&id, &text, &date, &name, &team, &vip, &userid)
		comments = append(comments, Comment{
			ID:   id,
			Text: text,
			User: User{
				ID:    userid,
				Name:  name,
				Vip:   vip,
				Staff: team,
			},
			Date: date,
		})
	}

	WriteResponse(200, comments, w)
}

// HandleComment takes care of POST on /comments which posts a comment
func HandleComment(w http.ResponseWriter, r *http.Request) {

}

// HandleVote takes care of POST on /vote which upvotes or downvotes a post
func HandleVote(w http.ResponseWriter, r *http.Request) {

}
