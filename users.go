package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user
type User struct {
	ID       *int64  `json:"id,omitempty"db:"id"`
	Username *string `json:"username,omitempty"db:"username"`
	FullName *string `json:"full_name,omitempty"db:"full_name"`
	Password *string `json:"password,omitempty"db:"password"`
}

// NewUser populates a User object
func NewUser(username, fullName, password string) *User {
	user := &User{}
	user.Username = &username
	user.FullName = &fullName
	user.Password = &password

	return user
}

func parseUserID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	vars := mux.Vars(r)
	idStr := vars["user_id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		sendBadReq(w, "invalid user id")
		return 0, false
	}
	exists, err := db().UserExists(id)
	if err != nil {
		sendInternalErr(w, err)
		return 0, false
	}
	if !exists {
		sendNotFound(w, fmt.Sprintf("user %d not found", id))
		return 0, false
	}

	return id, true
}

// CreateUserHandler handles POST /users
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	user := &User{}
	err := dec.Decode(&user)
	if err != nil {
		sendBadReq(w, "unable to decode request json")
		return
	}
	user.ID = nil

	if user.Username == nil {
		sendBadReq(w, "You need to provide a 'username'")
		return
	}
	if user.FullName == nil {
		sendBadReq(w, "You need to provide a 'full_name'")
		return
	}
	if user.Password == nil || len(*user.Password) == 0 {
		sendBadReq(w, "You need to provide a 'password'")
		return
	}

	// run the password through bcrypt
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(*user.Password), bcrypt.MaxCost)
	if err != nil {
		sendInternalErr(w, err)
		return
	}
	hashString := string(hashBytes)
	user.Password = &hashString

	id, err := db().CreateUser(user)
	if err != nil {
		sendInternalErr(w, err)
		return
	}
	user.ID = &id

	// the password hash should not be returned
	user.Password = nil

	sendSuccess(w, user)
}

// GetUserHandler handles GET /users/{user_id}
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	user, err := db().User(userID)
	if err != nil {
		sendInternalErr(w, err)
		return
	}

	// don't return the password hash
	user.Password = nil

	sendSuccess(w, user)
}

// EditUserHandler handles PUT /users/{user_id}
func EditUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	user, err := db().User(userID)
	if err != nil {
		sendInternalErr(w, err)
		return
	}

	oldPassword := user.Password

	dec := json.NewDecoder(r.Body)
	if err = dec.Decode(user); err != nil {
		sendBadReq(w, "unable to decode request json")
		return
	}

	// check if they're updating the password
	if *user.Password != *oldPassword {
		// we need to hash the new password, before writing it out to the db
		hashBytes, err := bcrypt.GenerateFromPassword([]byte(*user.Password), bcrypt.MaxCost)
		if err != nil {
			sendInternalErr(w, err)
			return
		}
		hashString := string(hashBytes)
		user.Password = &hashString
	}

	user.ID = &userID
	if err = db().EditUser(user); err != nil {
		sendInternalErr(w, err)
		return
	}

	// we don't want to send back the password
	user.Password = nil

	sendSuccess(w, user)
}
