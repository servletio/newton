package main

import (
	"bytes"
	"log"
	"os"
	"testing"
)

var bookmarkObject *Bookmark
var newUserID int64
var newAccessToken string
var hankHillContactID int64

const testUsername = "arash"
const testFullName = "Stinky Pete"
const testPassword = "a bad password"

var imageData = []byte{137, 80, 78, 71, 13, 10, 26, 10, 0, 0, 0, 13, 73, 72, 68, 82, 0, 0, 0, 27, 0, 0, 0, 27, 8, 4, 0, 0, 0, 39, 221, 60, 222, 0, 0, 0, 252, 73, 68, 65, 84, 120, 1, 237, 212, 161, 75, 107, 97, 28, 6, 224, 7, 150, 212, 102, 211, 164, 77, 133, 25, 214, 108, 155, 81, 48, 234, 13, 23, 46, 227, 150, 11, 6, 17, 220, 255, 225, 48, 136, 97, 147, 253, 21, 154, 134, 75, 22, 5, 141, 75, 227, 196, 45, 202, 101, 90, 6, 159, 240, 133, 3, 115, 158, 125, 101, 193, 224, 243, 134, 23, 62, 126, 239, 137, 199, 39, 171, 134, 246, 76, 171, 25, 90, 146, 112, 234, 213, 111, 151, 158, 12, 60, 186, 240, 199, 127, 127, 37, 149, 60, 8, 83, 233, 41, 73, 234, 8, 51, 185, 145, 240, 43, 158, 53, 101, 226, 64, 166, 25, 251, 200, 92, 47, 241, 168, 161, 47, 206, 244, 53, 98, 63, 155, 99, 93, 40, 204, 154, 66, 21, 89, 97, 42, 22, 165, 102, 100, 96, 71, 85, 40, 76, 85, 89, 102, 164, 42, 119, 39, 8, 174, 18, 179, 235, 216, 183, 114, 189, 248, 208, 73, 204, 58, 177, 123, 139, 159, 253, 204, 186, 241, 161, 157, 152, 181, 99, 119, 229, 78, 4, 19, 7, 137, 217, 161, 137, 224, 159, 28, 219, 54, 176, 59, 103, 86, 198, 166, 45, 95, 218, 87, 23, 98, 90, 234, 90, 66, 76, 93, 77, 74, 126, 42, 255, 4, 190, 219, 236, 61, 158, 30, 231, 63, 164, 55, 105, 56, 55, 214, 181, 140, 21, 247, 198, 206, 204, 248, 0, 255, 61, 90, 202, 148, 177, 201, 123, 0, 0, 0, 0, 73, 69, 78, 68, 174, 66, 96, 130}

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
	user := NewUser(testUsername, testFullName, testPassword)
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
	if *user.Username != testUsername {
		t.Fatalf("Wrong username: %s != %s", *user.Username, testUsername)
	}
	if user.FullName == nil {
		t.Fatal("Didn't retrieve full name")
	}
	if *user.FullName != testFullName {
		t.Fatalf("Wrong full name: %s != %s", *user.FullName, testFullName)
	}
	if user.Password == nil {
		t.Fatal("Didn't retrieve password")
	}
	if *user.Password != testPassword {
		t.Fatalf("Wrong password: %s != %s", *user.Password, testPassword)
	}
}

func TestUserExists(t *testing.T) {
	exists, err := db().UserExists(newUserID)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("Checking for the user returned false")
	}
}

func TestUserByUsername(t *testing.T) {
	user, err := db().UserByUsername(testUsername + "foo")
	if err != nil {
		t.Fatal(err)
	}
	if user != nil {
		t.Fatal("No user object was supposed to be returned for a bad username")
	}

	user, err = db().UserByUsername(testUsername)
	if err != nil {
		t.Fatal(err)
	}
	if user == nil {
		t.Fatal("User object was not returned")
	}
}

func TestEditUser(t *testing.T) {
	user, err := db().User(newUserID)
	if err != nil {
		t.Fatal(err)
	}
	if user == nil {
		t.Fatal("unable to retrieve user")
	}

	updatedPassword := "something"
	user.Password = &updatedPassword
	err = db().EditUser(user)
	if err != nil {
		t.Fatal(err)
	}

	updatedUser, err := db().User(newUserID)
	if err != nil {
		t.Fatal(err)
	}
	if *updatedUser.Password != updatedPassword {
		t.Fatalf("edit user failed: %s != %s", *updatedUser.Password, updatedPassword)
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
	retrieved, err := db().Bookmark(*bookmarkObject.ID, newUserID)
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
	retrieved, err := db().Bookmark(*bookmarkObject.ID, newUserID)
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
	err := db().DeleteBookmark(*bookmarkObject.ID, newUserID)
	if err != nil {
		t.Fatal(err)
	}

	retrieved, err := db().Bookmark(*bookmarkObject.ID, newUserID)
	if err != nil {
		t.Fatal(err)
	}
	if retrieved != nil {
		t.Fatal("A bookmark object was still returned after deleting it")
	}
}

func TestRetrieveBookmarks(t *testing.T) {
	bookmarks, err := db().Bookmarks(newUserID, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(bookmarks) < 1 {
		t.Fatal("Bookmarks() didn't return any bookmarks")
	}
}

func TestCreateSession(t *testing.T) {
	session := NewSession(newUserID)
	newAccessToken = *session.AccessToken

	sessionID, err := db().CreateSession(session)
	if err != nil {
		t.Fatal(err)
	}

	if sessionID < 1 {
		t.Fatal("invalid session id returned")
	}
}

func TestSessionByAccessToken(t *testing.T) {
	session, err := db().SessionByAccessToken(newAccessToken)
	if err != nil {
		t.Fatal(err)
	}

	if session == nil {
		t.Fatal("missing session object")
	}

	if session.AccessToken == nil {
		t.Fatal("access token was nil")
	}
	if *session.AccessToken != newAccessToken {
		t.Fatalf("access token did not match: %s != %s", *session.AccessToken, newAccessToken)
	}

	if session.UserID == nil {
		t.Fatal("user id of session was nil")
	}
	if *session.UserID != newUserID {
		t.Fatalf("retrieved user id did not match: %d != %d", *session.UserID, newUserID)
	}
}

func TestCreateContact(t *testing.T) {

	givenName := "Hank"
	familyName := "Hill"
	displayName := givenName + " " + familyName
	contact := &Contact{}
	contact.OwnerID = &newUserID

	contact.Name = &StructuredName{DisplayName: &displayName, GivenName: &givenName, FamilyName: &familyName}

	contact.Emails = append(contact.Emails, NewEmail("hank@gmail.com", "Home"))
	contact.Emails = append(contact.Emails, NewEmail("hank@stricklandpropane.com", "Work"))

	contact.Phones = append(contact.Phones, NewPhone("+1 214-555-1212", "Work"))
	contact.Phones = append(contact.Phones, NewPhone("469 555-1212", "Mobile"))

	contact.IMHandles = append(contact.IMHandles, NewIMHandle("hank@gmail.com", "gtalk", "Home"))
	contact.IMHandles = append(contact.IMHandles, NewIMHandle("hank@stricklandpropane.com", "xmpp", "Work"))

	contact.Org = &Organization{}
	company := "Strickland Propane"
	contact.Org.Company = &company
	title := "Assistant Manager"
	contact.Org.Title = &title
	department := "Sales"
	contact.Org.Department = &department
	jobDescription := "Selling propane and propane accessories"
	contact.Org.JobDescription = &jobDescription
	officeLocation := "In the back"
	contact.Org.OfficeLocation = &officeLocation

	contact.Relations = append(contact.Relations, NewRelation("Peggy", "wife"))
	contact.Relations = append(contact.Relations, NewRelation("Bobby", "son"))
	contact.Relations = append(contact.Relations, NewRelation("Luanne", "niece"))
	contact.Relations = append(contact.Relations, NewRelation("Junichiro", "half-brother"))

	workAddress := NewUSAAddress("135 Los Gatos Road", "Arlen", "Texas", "12345", "work")
	homeAddress := NewUSAAddress("123 Don't Know St", "Arlen", "Texas", "12345", "home")
	contact.PostalAddresses = append(contact.PostalAddresses, workAddress, homeAddress)

	contact.Websites = append(contact.Websites, NewWebsite("https://stricklandpropane.com", "work"))
	contact.Websites = append(contact.Websites, NewWebsite("https://hillfamily.com/hank", "personal"))

	contact.Events = append(contact.Events, NewEvent("April 19, 1957", "birthday"))

	var err error
	hankHillContactID, err = db().CreateContact(contact)
	if err != nil {
		t.Fatal(err)
	}
	if hankHillContactID < 1 {
		t.Fatalf("did not get a valid id: %d", hankHillContactID)
	}
}

func TestSetContactPhoto(t *testing.T) {
	err := db().SetContactPhoto(hankHillContactID, imageData)
	if err != nil {
		t.Fatal(err)
	}

	retrievedData, err := db().ContactPhoto(hankHillContactID)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(retrievedData, imageData) {
		t.Fatal("retrieved image data does not match the original data put into the database")
	}

	err = db().SetContactPhoto(hankHillContactID, nil)
	if err != nil {
		t.Fatal(err)
	}

	retrievedData, err = db().ContactPhoto(hankHillContactID)
	if err != nil {
		t.Fatal(err)
	}
	if retrievedData != nil {
		t.Fatal("retrieved image should be nil after deletion")
	}
}
