package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Session contains the fields used in managing a user's authentication
type Session struct {
	ID           *int64     `json:"id,omitempty"db:"id"`
	AccessToken  *string    `json:"access_token,omitempty"db:"access_token"`
	UserID       *int64     `json:"user_id,omitempty"db:"user_id"`
	CreationDate *time.Time `json:"creation_date,omitempty"db:"creation_date"`
}

// NewSession creates a session, and generates an access token and creation date
func NewSession(userID int64) *Session {
	session := &Session{UserID: &userID}
	now := time.Now()
	session.CreationDate = &now
	token := randAlphaNum(32)
	session.AccessToken = &token

	return session
}

func authenticate(w http.ResponseWriter, r *http.Request) (int64, bool) {
	args := r.URL.Query()
	accessToken := args.Get("access_token")
	token := strings.TrimSpace(accessToken)
	if token == "" {
		sendUnauthorized(w, "you need to be logged in to continue")
		return 0, false
	}

	session, err := db().SessionByAccessToken(token)
	if err != nil {
		sendInternalErr(w, err)
		return 0, false
	}

	if session == nil {
		sendUnauthorized(w, "you need to be logged in to continue")
		return 0, false
	}

	return *session.UserID, true
}

// CreateSessionHandler handles POST /sessions
func CreateSessionHandler(w http.ResponseWriter, r *http.Request) {
	userAndPass := struct {
		Username *string `json:"username,omitempty"`
		Password *string `json:"password,omitempty"`
	}{}

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&userAndPass)
	if err != nil {
		sendBadReq(w, "unable to decode request json")
		return
	}

	if userAndPass.Username == nil || len(*userAndPass.Username) == 0 {
		sendBadReq(w, "you need to specify a 'user_name'")
		return
	}
	if userAndPass.Password == nil {
		sendBadReq(w, "you need to specify a 'password'")
		return
	}

	user, err := db().UserByUsername(*userAndPass.Username)
	if err != nil {
		sendInternalErr(w, err)
		return
	}
	if user == nil {
		sendUnauthorized(w, "incorrect 'username' and/or 'password'")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(*userAndPass.Password))
	if err != nil {
		sendUnauthorized(w, "incorrect 'username' and/or 'password'")
		return
	}

	// make a new session, persist it, then return it
	session := NewSession(*user.ID)
	sessionID, err := db().CreateSession(session)
	if err != nil {
		sendInternalErr(w, err)
		return
	}
	session.ID = &sessionID

	sendSuccess(w, session)
}
