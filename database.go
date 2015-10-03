package main

var gDatabase NewtonDB

// NewtonDB is an abstraction of all the methods necessary for a database provider to implement
type NewtonDB interface {
	Bookmark(id int64) (*Bookmark, error)
	BookmarkExists(id int64) (bool, error)
	Bookmarks(userID int64, pageSize int, page int) ([]*Bookmark, error)
	CreateBookmark(bookmark *Bookmark) (int64, error)
	DeleteBookmark(id int64) error
	EditBookmark(bookmark *Bookmark) error

	User(id int64) (*User, error)
	UserExists(id int64) (bool, error)
	UserByUsername(username string) (*User, error)
	CreateUser(user *User) (int64, error)
	EditUser(user *User) error

	CreateSession(session *Session) (int64, error)
	SessionByAccessToken(token string) (*Session, error)

	CreateContact(contact *Contact) (int64, error)
}

// InitDB initializes the database that backs the API
func InitDB(connectInfo string) error {
	var err error
	gDatabase, err = NewMariaDB(connectInfo)
	return err
}

func db() NewtonDB {
	return gDatabase
}
