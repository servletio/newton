package main

import (
	"log"
	"os"
	"testing"
)

var bookmarkObject *Bookmark
var newUserID int64

func TestInstantiateDatabase(t *testing.T) {
	dsn := os.Getenv("SQL_DB")
	if dsn == "" {
		log.Fatal("You need to specify the SQL_DB environment variable to instantiate the database.")
	}

	if err := InitDB(dsn); err != nil {
		log.Fatal(err)
	}
}

func TestCreateUser(t *testing.T) {
	user := NewUser("arash", "Stinky Pete", "a bad password")
	userID, err := db().CreateUser(user)
	if err != nil {
		t.Fatal(err)
	}
	user.ID = &userID
	newUserID = userID
}

func TestGetUser(t *testing.T) {
	user, err := db().User(newUserID)
	if err != nil {
		t.Fatal(err)
	}
	if user.Username == nil {
		t.Fatal("Didn't retrieve username")
	}
	if user.FullName == nil {
		t.Fatal("Didn't retrieve full name")
	}
	if user.Password == nil {
		t.Fatal("Didn't retrieve password")
	}
}

func TestCreateBookmark(t *testing.T) {
	bookmarkObject = NewBookmark("http://ara.sh", "Official site of Arash Payan", newUserID)
	firstID, err := db().CreateBookmark(bookmarkObject)
	if err != nil {
		t.Fatal(err)
	}
	bookmarkObject.ID = &firstID

	secondID, err := db().CreateBookmark(NewBookmark("https://arashpayan.com", "The site of Arash Payan", newUserID))
	if err != nil {
		t.Fatal(err)
	}

	if firstID == secondID {
		t.Fatalf("Bookmark ids must be unique")
	}
}

func TestRetrieveBookmark(t *testing.T) {
	retrieved, err := db().Bookmark(*bookmarkObject.ID)
	if err != nil {
		t.Fatal(err)
	}

	if retrieved == nil {
		t.Fatal("The returned bookmark was nil")
	}
	if *retrieved.ID != *bookmarkObject.ID {
		t.Fatal("Bookmark.ID did not match after retrieval")
	}
	if *retrieved.URL != *bookmarkObject.URL {
		t.Fatal("Bookmark.URL did not match after retrieval")
	}
	if *retrieved.Title != *bookmarkObject.Title {
		t.Fatal("Bookmark.Title did not match after retrieval")
	}
	if retrieved.OwnerID != nil && *retrieved.OwnerID != *bookmarkObject.OwnerID {
		t.Fatal("Bookmark.UserID did not match after retrieval")
	}
}

func TestEditBookmark(t *testing.T) {
	url := "https://news.ycombinator.com"
	title := "Hacker News"
	bookmarkObject.URL = &url
	bookmarkObject.Title = &title

	err := db().EditBookmark(bookmarkObject)
	if err != nil {
		t.Fatal(err)
	}

	// retrieve it and make sure the fields match
	retrieved, err := db().Bookmark(*bookmarkObject.ID)
	if err != nil {
		t.Fatalf("unable to retrieve bookmark while verifying that the edit stuck")
	}

	if retrieved == nil {
		t.Fatal("The returned bookmark was nil after editing")
	}
	if *retrieved.ID != *bookmarkObject.ID {
		t.Fatalf("Bookmark.ID did not match after editing: original (%d) != retrieved (%d)", *retrieved.ID, *bookmarkObject.ID)
	}
	if *retrieved.URL != *bookmarkObject.URL {
		t.Fatal("Bookmark.URL did not match after editing")
	}
	if *retrieved.Title != *bookmarkObject.Title {
		t.Fatal("Bookmark.Title did not match after editing")
	}
}

func TestDeleteBookmark(t *testing.T) {
	err := db().DeleteBookmark(*bookmarkObject.ID)
	if err != nil {
		t.Fatal(err)
	}

	retrieved, err := db().Bookmark(*bookmarkObject.ID)
	if err != nil {
		t.Fatal(err)
	}
	if retrieved != nil {
		t.Fatal("A bookmark object was still returned after deleting it")
	}
}

func TestRetrieveBookmarks(t *testing.T) {
	bookmarks, err := db().Bookmarks(nil, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(bookmarks) < 1 {
		t.Fatal("Bookmarks() didn't return any bookmarks")
	}

	log.Printf("bookmarks: %v", bookmarks)
}
