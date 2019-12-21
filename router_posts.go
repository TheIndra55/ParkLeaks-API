package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// HandlePosts function serves /posts which returns all posts
func HandlePosts(w http.ResponseWriter, r *http.Request) {
	rows, err := Db.Query("SELECT posts.id, posts.titel, posts.text, posts.images, posts.verified, posts.date, posts.views, namen.name, namen.id as userid, namen.vip, " +
		"namen.team FROM `posts` INNER JOIN `namen` ON posts.address=namen.address WHERE `public` = 1 ORDER BY `id` DESC")
	defer rows.Close()

	if err != nil {
		w.WriteHeader(500)
		WriteErrors(500, []string{"An internal server error occured", "Something went wrong while retrieving the data"}, w)
		return
	}

	var posts []Post
	for rows.Next() {
		var (
			id, views, userid   int
			title, text, name   string
			images              string
			verified, vip, team bool
			date                time.Time
		)

		rows.Scan(&id, &title, &text, &images, &verified, &date, &views, &name, &userid, &vip, &team)
		posts = append(posts, Post{
			ID:       id,
			Title:    title,
			Text:     text,
			Verified: verified,
			Date:     date,
			Images:   Split(images, ","),
			Stats:    Stats{Views: views},
			User: User{
				ID:    &userid,
				Name:  name,
				Vip:   vip,
				Staff: team,
			},
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
	rows, err := Db.Query("SELECT posts.id, posts.titel, posts.text, posts.images, posts.verified, posts.date, posts.views, namen.name, namen.id as userid, namen.vip, "+
		"namen.team FROM `posts` INNER JOIN `namen` ON posts.address=namen.address WHERE posts.id = ?", mux.Vars(r)["post"])

	if err != nil {
		w.WriteHeader(500)
		WriteErrors(500, []string{"An internal server error occured", "Something went wrong while retrieving the data"}, w)
		return
	}

	var (
		id, views, userid, vote int
		title, text, name       string
		images                  string
		verified, vip, team     bool
		date                    time.Time
	)

	rows.Next()
	rows.Scan(&id, &title, &text, &images, &verified, &date, &views, &name, &userid, &vip, &team)
	rows.Close()

	ip := r.Header.Get("CF-Connecting-IP")

	// get user voted state
	err = Db.QueryRow("SELECT `action` FROM `votes` WHERE `postid` = ? and `address` = ?", mux.Vars(r)["post"], ip).Scan(&vote)

	// TODO: refactor
	if vote == 0 {
		vote = -1
	}

	// if error just assume no vote so 0
	if err != nil {
		vote = 0
	}

	WriteResponse(200, Post{
		ID:       id,
		Title:    title,
		Text:     text,
		Verified: verified,
		Date:     date,
		Images:   Split(images, ","),
		Stats: Stats{
			Score: CountVotes(id),
			Views: views,
		},
		User: User{
			ID:    &userid,
			Name:  name,
			Vip:   vip,
			Staff: team,
		},
		Vote: vote,
	}, w)
}

// HandleComments services comments for a single post
func HandleComments(w http.ResponseWriter, r *http.Request) {
	rows, err := Db.Query("SELECT reacties.id, reacties.text, reacties.date, namen.name, namen.team, namen.vip, namen.id as `userid` FROM `reacties` "+
		"INNER JOIN `namen` ON reacties.address=namen.address WHERE `postid` = ? ORDER BY `id` DESC", mux.Vars(r)["post"])
	defer rows.Close()

	if err != nil {
		w.WriteHeader(500)
		WriteErrors(500, []string{"An internal server error occured", "Something went wrong while retrieving the data"}, w)
		return
	}

	var comments []Comment
	for rows.Next() {
		var (
			id, userid int
			text, name string
			team, vip  bool
			date       time.Time
		)

		rows.Scan(&id, &text, &date, &name, &team, &vip, &userid)
		comments = append(comments, Comment{
			ID:   id,
			Text: text,
			User: &User{
				ID:    &userid,
				Name:  name,
				Vip:   vip,
				Staff: team,
			},
			Date: date,
		})
	}

	WriteResponse(200, comments, w)
}

// CountVotes returns the score of a post
func CountVotes(post int) int {
	rows, _ := Db.Query("SELECT COUNT(CASE WHEN `action` = 0 THEN 1 END), COUNT(CASE WHEN `action` = 1 THEN 1 END) FROM `votes` WHERE `postid` = ?", post)
	defer rows.Close()

	var down, up int
	rows.Next()
	rows.Scan(&down, &up)

	return (0 + up) - down
}
