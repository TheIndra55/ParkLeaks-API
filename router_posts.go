package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// HandlePosts function serves /posts which returns all posts
func HandlePosts(w http.ResponseWriter, r *http.Request) {
	var (
		page int
		err  error
		rows *sql.Rows
	)
	param, ok := r.URL.Query()["page"]

	if !ok || len(param) == 0 {
		// if param doesn't exist default to page 0
		page = 0
	} else if page, err = strconv.Atoi(param[0]); err != nil {
		w.WriteHeader(400)
		WriteError(400, "Invalid value for parameter 'page', should be an integer", w)
		return
	}

	// hardcoded limit of 8
	// TODO: user-defined limit
	limit := 8
	// calculate offset and pass to query
	offset := page * limit

	rows, err = Db.Query("SELECT posts.id, posts.titel, posts.text, posts.images, posts.verified, posts.date, posts.views, namen.name, namen.id as userid, namen.vip, "+
		"namen.team FROM `posts` INNER JOIN `namen` ON posts.address=namen.address WHERE `public` = 1 ORDER BY `id` DESC LIMIT ? OFFSET ?", limit, offset)
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
			thumbnails          []string
		)

		rows.Scan(&id, &title, &text, &images, &verified, &date, &views, &name, &userid, &vip, &team)

		splitted := Split(images, ",")

		// return array with only 1 thumbnail
		if len(splitted) == 0 {
			thumbnails = []string{}
		} else {
			thumbnails = []string{
				GenerateThumbnail(fmt.Sprintf("https://parkleaks.nl/%s", splitted[0]), 300),
			}
		}

		posts = append(posts, Post{
			ID:       id,
			Title:    title,
			Text:     text,
			Verified: verified,
			Date:     date,
			Images:   thumbnails,
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

// GenerateThumbnail generates the thumbnail url for an image url and scale
func GenerateThumbnail(image string, scale int) string {
	hash := Hash(image)
	hex := hex.EncodeToString([]byte(image))

	url := url.URL{
		Scheme:   "https",
		Host:     "user.blazor.nl",
		Path:     path.Join("scale", hash, hex),
		RawQuery: fmt.Sprintf("scale=%d&signature=%s", scale, Hash(strconv.Itoa(scale))),
	}

	return url.String()
}

// Hash returns a new HMAC hash of the given value using key from ENV
func Hash(value string) string {
	key := []byte(os.Getenv("CAMO_KEY"))
	hash := hmac.New(sha1.New, key)

	hash.Write([]byte(value))
	return hex.EncodeToString(hash.Sum(nil))
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
	
	imageList := Split(images, ",")
	
	/*for i, val := range imageList {
		prefix := "https://parkleaks.nl/"
		imageList[i] = prefix+string(val)
	}*/

	WriteResponse(200, Post{
		ID:       id,
		Title:    title,
		Text:     text,
		Verified: verified,
		Date:     date,
		Images:   imageList,
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
