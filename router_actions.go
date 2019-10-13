package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

// PayloadVote is the vote payload
type PayloadVote struct {
	Action int `json:"action"`
}

// PayloadComment is the payload for posting a comment
type PayloadComment struct {
	Comment  string `json:"comment"`
	Response string `json:"captcha"`
}

// RecaptchaResponse is serialized version of reCAPTCHA response
type RecaptchaResponse struct {
	Success bool `json:"success"`
}

// HandleComment takes care of POST on /comments which posts a comment
func HandleComment(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		WriteError(400, "The body couldn't be read", w)
		return
	}

	var comment PayloadComment

	err = json.Unmarshal(body, &comment)
	if err != nil {
		w.WriteHeader(400)
		WriteError(400, "The body was an invalid JSON payload", w)
		return
	}

	if len(comment.Comment) <= 5 {
		w.WriteHeader(400)
		WriteError(400, "Field 'comment' must be longer than 5 characters", w)
		return
	}

	if len(comment.Comment) > 600 {
		w.WriteHeader(400)
		WriteError(400, "Field 'comment' must not be longer than 600 characters", w)
		return
	}

	//ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	ip := r.Header.Get("CF-Connecting-IP")
	post := mux.Vars(r)["post"]

	passed, err := CheckCaptcha(comment.Response, ip)
	if err != nil {
		w.WriteHeader(500)
		WriteError(500, "Something got wrong while contacting the reCAPTCHA provider", w)
		return
	}

	if !passed {
		w.WriteHeader(403)
		WriteError(403, "Invalid captcha", w)
		return
	}

	exists, err := DoesExists(User{Address: ip})
	if err != nil {
		w.WriteHeader(500)
		WriteErrors(500, []string{"An internal server error occured", "Something went wrong while checking if user exists"}, w)
		return
	}

	if !exists {
		// generate a new account if user doesn't have account yet
		GenerateAccount(ip)
	}

	rows, _ := Db.Query("INSERT INTO `reacties` (postid, text, date, address) VALUES (?, ?, NOW(), ?)", post, comment.Comment, ip)
	defer rows.Close()

	WriteResponse(200, nil, w)
}

// CheckCaptcha checks a captcha token at the reCAPTCHA provider
func CheckCaptcha(response string, ip string) (bool, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://www.google.com/recaptcha/api/siteverify", nil)

	q := req.URL.Query()
	q.Add("secret", os.Getenv("RECAPTCHA_SECRET"))
	q.Add("response", response)
	q.Add("remoteip", ip)
	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)
	if err != nil {
		return false, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	var captcharesponse RecaptchaResponse
	err = json.Unmarshal(body, &captcharesponse)
	if err != nil {
		return false, err
	}

	return captcharesponse.Success, nil
}

// HandleVote takes care of POST on /vote which upvotes or downvotes a post
func HandleVote(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		WriteError(400, "The body couldn't be read", w)
		return
	}

	var vote PayloadVote
	err = json.Unmarshal(body, &vote)
	if err != nil {
		w.WriteHeader(400)
		WriteError(400, "The body was an invalid JSON payload", w)
		return
	}

	//address, _, _ := net.SplitHostPort(r.RemoteAddr)
	address := r.Header.Get("CF-Connecting-IP")
	post := mux.Vars(r)["post"]

	switch vote.Action {
	case 0:
		// remove vote
		docs, _ := Db.Query("DELETE FROM `votes` WHERE `postid` = ? AND `address` = ?", post, address)
		docs.Close()
	case -1, 1:
		// vote
		// check if user already voted
		rows, err := Db.Query("SELECT COUNT(*) FROM `votes` WHERE `postid` = ? AND `address` = ?", post, address)
		defer rows.Close()

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
			rows, _ := Db.Query("UPDATE `votes` SET `action` = ? WHERE `postid` = ? AND `address` = ?", vote.Action, post, address)
			defer rows.Close()
		} else {
			// insert
			rows, _ := Db.Query("INSERT INTO `votes` (`postid`, `address`, `action`) VALUES(?, ?, ?)", post, address, vote.Action)
			defer rows.Close()
		}
	default:
		// invalid action
		w.WriteHeader(400)
		WriteError(400, "An invalid action was specified", w)
		return
	}

	// return updated vote count
	id, _ := strconv.Atoi(post)
	WriteResponse(200, struct {
		Score int `json:"score"`
	}{CountVotes(id)}, w)
}
