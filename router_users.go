package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

// UserResponse is used as data in user lookup endpoint
type UserResponse struct {
	User     User      `json:"user"`
	Comments []Comment `json:"comments"`
}

// HandleUser handles a user lookup endpoint
func HandleUser(w http.ResponseWriter, r *http.Request) {
	rows, err := Db.Query("SELECT `id`, `name`, `team`, `vip`, `address` FROM `namen` WHERE `id` = ?", mux.Vars(r)["user"])
	if err != nil {
		w.WriteHeader(500)
		WriteErrors(500, []string{"An internal server error occured", "Something went wrong while retrieving the data"}, w)
		return
	}

	var (
		id            int
		name, address string
		staff, vip    bool
	)
	rows.Next()
	rows.Scan(&id, &name, &staff, &vip, &address)

	user := User{
		ID:      id,
		Name:    name,
		Staff:   staff,
		Vip:     vip,
		Address: address,
	}
	WriteResponse(200, UserResponse{User: user, Comments: GetComments(user)}, w)
}

// GetComments returns all comments by a user
func GetComments(user User) []Comment {
	rows, _ := Db.Query("SELECT `id`, `text`, `date` FROM `reacties` WHERE `address` = ? ORDER BY `id` DESC", user.Address)
	var comments []Comment

	for rows.Next() {
		var (
			id         int
			text, date string
		)
		rows.Scan(&id, &text, &date)
		comments = append(comments, Comment{
			ID:   id,
			Text: text,
			Date: date,
		})
	}

	return comments
}
