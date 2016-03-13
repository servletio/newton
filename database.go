package main

var gDatabase NewtonDB

// NewtonDB is an abstraction of all the methods necessary for a database provider to implement
type NewtonDB interface {
	Bookmark(bookmarkID, ownerID int64) (*Bookmark, error)
	BookmarkExists(id int64) (bool, error)
	Bookmarks(ownerID int64, pageSize int, page int) ([]*Bookmark, error)
	CreateBookmark(bookmark *Bookmark) (int64, error)
	DeleteBookmark(bookmarkID, ownerID int64) error
	EditBookmark(bookmark *Bookmark) error

	User(id int64) (*User, error)
	UserExists(id int64) (bool, error)
	UserByUsername(username string) (*User, error)
	CreateUser(user *User) (int64, error)
	EditUser(user *User) error

	CreateSession(session *Session) (int64, error)
	SessionByAccessToken(token string) (*Session, error)

	CreateContact(contact *Contact) (int64, error)
	ContactExists(id int64) (bool, error)
	Contact(contactID, ownerID int64) (*Contact, error)
	Contacts(ownerID int64) ([]*Contact, error)
	DeleteContact(contactID, ownerID int64) error
	SetContactPhoto(contactID int64, photo []byte) error
	ContactPhoto(contactID int64) ([]byte, error)
	ContactOwner(contactID int64) (int64, error)

	AddLocationRecord(locRec *LocationRecord) error
}

// InitDB initializes the database that backs the API
func InitDB(connectInfo string) error {
	var err error
	// gDatabase, err = NewMariaDB(connectInfo)
	gDatabase, err = NewSQLiteDB(connectInfo)
	return err
}

func db() NewtonDB {
	return gDatabase
}
