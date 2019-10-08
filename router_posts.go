package main

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// HandlePosts function serves /posts which returns all posts
func HandlePosts(w http.ResponseWriter, r *http.Request) {
	rows, err := Db.Query("SELECT `id`, `titel`, `text`, `images`, `verified`, `date` FROM `posts` WHERE `public` = 1 ORDER BY `id` DESC")
	if err != nil {
		w.WriteHeader(500)
		WriteErrors(500, []string{"An internal server error occured", "Something went wrong while retrieving the data"}, w)
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
		WriteErrors(500, []string{"An internal server error occured", "Something went wrong while retrieving the data"}, w)
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
		Stats: Stats{
			Score: CountVotes(id),
		},
	}, w)
}

// HandleComments services comments for a single post
func HandleComments(w http.ResponseWriter, r *http.Request) {
	rows, err := Db.Query("SELECT reacties.id, reacties.text, reacties.date, namen.name, namen.team, namen.vip, namen.id as `userid` FROM `reacties` "+
		"INNER JOIN `namen` ON reacties.address=namen.address WHERE `postid` = ? ORDER BY `id` DESC", mux.Vars(r)["post"])

	if err != nil {
		w.WriteHeader(500)
		WriteErrors(500, []string{"An internal server error occured", "Something went wrong while retrieving the data"}, w)
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
			User: &User{
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

// Vote is the vote payload
type Vote struct {
	Action int `json:"action"`
}

// HandleVote takes care of POST on /vote which upvotes or downvotes a post
func HandleVote(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		WriteError(400, "The body couldn't be read", w)
		return
	}

	var vote Vote
	err = json.Unmarshal(body, &vote)
	if err != nil {
		w.WriteHeader(400)
		WriteError(400, "The body was an invalid JSON payload", w)
		return
	}

	address, _, _ := net.SplitHostPort(r.RemoteAddr)
	post := mux.Vars(r)["post"]

	switch vote.Action {
	case 0:
		// remove vote
		Db.Query("DELETE FROM `votes` WHERE `postid` = ? AND `address` = ?", post, address)
	case -1, 1:
		// vote
		// check if user already voted
		rows, err := Db.Query("SELECT COUNT(*) FROM `votes` WHERE `postid` = ? AND `address` = ?", post, address)
		if err != nil {
			w.WriteHeader(500)
			WriteErrors(500, []string{"An internal server error occured", "Something went wrong while retrieving the data"}, w)
			return
		}

		// dirty hack as -1 is 0 in vote column
		if vote.Action == -1 {
			vote.Action = 0
		}

		var count int
		rows.Next()
		rows.Scan(&count)

		if count > 0 {
			// already voted update
			Db.Query("UPDATE `votes` SET `action` = ? WHERE `postid` = ? AND `address` = ?", vote.Action, post, address)
		} else {
			// insert
			Db.Query("INSERT INTO `votes` (`postid`, `address`, `action`) VALUES(?, ?, ?)", post, address, vote.Action)
		}
	default:
		// invalid action
		w.WriteHeader(400)
		WriteError(400, "An invalid action was specified", w)
		return
	}

	// return updated vote count
	id, _ := strconv.Atoi(post)
	WriteResponse(200, CountVotes(id), w)
}

// CountVotes returns the score of a post
func CountVotes(post int) int {
	rows, _ := Db.Query("SELECT COUNT(CASE WHEN `action` = 0 THEN 1 END), COUNT(CASE WHEN `action` = 1 THEN 1 END) FROM `votes` WHERE `postid` = ?", post)

	var down, up int
	rows.Next()
	rows.Scan(&down, &up)

	return (0 + up) - down
}
