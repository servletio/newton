package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Bookmark represents a bookmark
type Bookmark struct {
	ID      *int64  `json:"id,omitempty"db:"id"`
	URL     *string `json:"url,omitempty"db:"url"`
	Title   *string `json:"title,omitempty"db:"title"`
	OwnerID *int64  `json:"owner_id,omitempty"db:"owner_id"`
}

func (b Bookmark) String() string {
	buf, _ := json.MarshalIndent(b, "", " ")
	return string(buf)
}

// NewBookmark creates a new bookmark object (in-memory only)
func NewBookmark(url, title string, ownerID int64) *Bookmark {
	b := &Bookmark{}
	b.URL = &url
	b.Title = &title
	b.OwnerID = &ownerID

	return b
}

func parseBookmarkID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	vars := mux.Vars(r)
	idStr := vars["bookmark_id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		sendBadReq(w, "invalid bookmark id")
		return 0, false
	}
	exists, err := db().BookmarkExists(id)
	if err != nil {
		sendInternalErr(w, err)
		return 0, false
	}
	if !exists {
		sendNotFound(w, fmt.Sprintf("bookmark %d not found", id))
		return 0, false
	}

	return id, true
}

// CreateBookmarkHandler handles POST /bookmarks
func CreateBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	bookmark := &Bookmark{}
	err := decoder.Decode(bookmark)
	if err != nil {
		sendBadReq(w, "unable to decode the request json")
		return
	}
	bookmark.ID = nil

	if bookmark.URL == nil {
		sendBadReq(w, "You need to provide a 'url'")
		return
	}

	id, err := db().CreateBookmark(bookmark)
	if err != nil {
		sendInternalErr(w, err)
		return
	}
	bookmark.ID = &id
	sendSuccess(w, bookmark)
}

// GetBookmarksHandler handles GET /bookmarks
func GetBookmarksHandler(w http.ResponseWriter, r *http.Request) {
	page, pageSize, err := pageAndSize(r.URL.Query(), 10)
	if err != nil {
		sendBadReq(w, err.Error())
		return
	}

	bookmarks, err := db().Bookmarks(nil, pageSize, page)
	sendSuccess(w, bookmarks)
}

// GetBookmarkHandler handles GET /bookmarks/{bookmark_id}
func GetBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	bookmarkID, ok := parseBookmarkID(w, r)
	if !ok {
		return
	}

	bookmark, err := db().Bookmark(bookmarkID)
	if err != nil {
		sendInternalErr(w, err)
		return
	}

	sendSuccess(w, bookmark)
}

// EditBookmarkHandler handles PUT /bookmarks/{bookmark_id}
func EditBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	bookmarkID, ok := parseBookmarkID(w, r)
	if !ok {
		return
	}

	bookmark, err := db().Bookmark(bookmarkID)
	if err != nil {
		sendInternalErr(w, err)
		return
	}
	bookmarkOwnerID := bookmark.OwnerID

	dec := json.NewDecoder(r.Body)
	if err = dec.Decode(bookmark); err != nil {
		sendBadReq(w, "unable to decode the request json")
		return
	}

	bookmark.ID = &bookmarkID
	bookmark.OwnerID = bookmarkOwnerID
	if err = db().EditBookmark(bookmark); err != nil {
		sendInternalErr(w, err)
		return
	}

	sendSuccess(w, bookmark)
}

// DeleteBookmarkHandler handles DELETE /bookmarks/{bookmark_id}
func DeleteBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	bookmarkID, ok := parseBookmarkID(w, r)
	if !ok {
		return
	}

	if err := db().DeleteBookmark(bookmarkID); err != nil {
		sendInternalErr(w, err)
		return
	}

	sendSuccess(w, nil)
}
