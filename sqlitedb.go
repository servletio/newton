package main

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	_ "github.com/mattn/go-sqlite3"
)

// DropAllMariaDBTables is just useful when testing
const DropAllMariaDBTables = `
DROP TABLE bookmarks,
           contacts,
		   contacts_name,
		   contacts_emails,
		   contacts_events,
		   ontacts_im_accounts,
		   contacts_organization,
		   contacts_phones,
		   contacts_photo,
		   contacts_postal_addresses,
		   contacts_relations,
		   contacts_websites,
		   database_version,
		   essions,
		   users`

// CreateTableDatabaseVersion is the statement to create a table that tracks the current schema version
const CreateTableDatabaseVersion = `
CREATE TABLE IF NOT EXISTS database_version (id INTEGER PRIMARY KEY NOT NULL,
                                             version INT NOT NULL DEFAULT 0)`

// CreateTableBookmarks is the statement to create the bookmarks table
const CreateTableBookmarks = `
CREATE TABLE IF NOT EXISTS bookmarks (id INTEGER PRIMARY KEY,
                                      url TEXT NOT NULL,
									  title TEXT NOT NULL DEFAULT "title",
									  owner_id INTEGER NOT NULL)`

// CreateTableUsers is the statement to create the users table
const CreateTableUsers = `
CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY NOT NULL,
                                  username TEXT NOT NULL,
								  full_name TEXT NOT NULL,
								  password TEXT NOT NULL)`

// CreateTableSessions is the statement to create the sessions table
const CreateTableSessions = `
CREATE TABLE IF NOT EXISTS sessions (id INTEGER PRIMARY KEY NOT NULL,
                                     access_token TEXT NOT NULL,
									 user_id INTEGER NOT NULL,
									 creation_date TIMESTAMP NOT NULL)`

// CreateTableContacts is the statement to create the contacts table
const CreateTableContacts = `
CREATE TABLE IF NOT EXISTS contacts (id INTEGER PRIMARY KEY NOT NULL,
	                                 nickname TEXT,
									 note TEXT,
                                     owner_id INTEGER NOT NULL)`

// CreateTableContactsName is the statement to create the table for storing a contact's name
const CreateTableContactsName = `
CREATE TABLE IF NOT EXISTS contacts_name (contact_id INTEGER PRIMARY KEY NOT NULL,
                                          display_name TEXT,
										  prefix TEXT,
										  given_name TEXT,
										  middle_name TEXT,
										  family_name TEXT,
										  suffix TEXT,
										  phonetic_given_name TEXT,
										  phonetic_middle_name TEXT,
										  phonetic_family_name TEXT)`

// CreateTableContactsEmails creates the table for storing a contact's emails
const CreateTableContactsEmails = `
CREATE TABLE IF NOT EXISTS contacts_emails (id INTEGER PRIMARY KEY NOT NULL,
                                            contact_id INTEGER NOT NULL,
											address TEXT,
											type INTEGER,
											label TEXT)`

// CreateTableContactsPhones creates the table for storing a contact's phone numbers
const CreateTableContactsPhones = `
CREATE TABLE IF NOT EXISTS contacts_phones (id INTEGER PRIMARY KEY NOT NULL,
                                            contact_id INTEGER NOT NULL,
											number TEXT,
											type INTEGER,
											label TEXT)`

// CreateTableContactsIMAccounts creates the table for storing a contact's IM handles
const CreateTableContactsIMAccounts = `
CREATE TABLE IF NOT EXISTS contacts_im_accounts (id INTEGER PRIMARY KEY NOT NULL,
                                                 contact_id INTEGER NOT NULL,
												 handle TEXT,
												 type INTEGER,
												 label TEXT,
												 protocol INTEGER,
												 custom_protocol TEXT)`

// CreateTableContactsOrganization creates the table for storing a contact's organization/association details
const CreateTableContactsOrganization = `
CREATE TABLE IF NOT EXISTS contacts_organization (contact_id INTEGER PRIMARY KEY NOT NULL,
                                                  company TEXT,
												  title TEXT)`

// CreateTableContactsRelations creates the table for storing a contact's relations (spouse, children, etc.)
const CreateTableContactsRelations = `
CREATE TABLE IF NOT EXISTS contacts_relations (id INTEGER PRIMARY KEY NOT NULL,
                                               contact_id INTEGER NOT NULL,
											   name TEXT,
											   type TEXT)`

// CreateTableContactsPostalAddresses creates the table for storing a contact's postal addresses
const CreateTableContactsPostalAddresses = `
CREATE TABLE IF NOT EXISTS contacts_postal_addresses (id INTEGER PRIMARY KEY NOT NULL,
                                                      contact_id INTEGER NOT NULL,
													  street TEXT,
													  po_box TEXT,
													  neighborhood TEXT,
													  city TEXT,
													  region TEXT,
													  post_code TEXT,
													  country TEXT,
													  type INTEGER,
													  label TEXT)`

// CreateTableContactsWebsites creates the table for storing a contact's websites
const CreateTableContactsWebsites = `
CREATE TABLE IF NOT EXISTS contacts_websites (id INTEGER PRIMARY KEY NOT NULL,
                                              contact_id INTEGER NOT NULL,
											  address TEXT,
											  type TEXT)`

// CreateTableContactsEvents creates the table for storing a contact's events
const CreateTableContactsEvents = `
CREATE TABLE IF NOT EXISTS contacts_events (id INTEGER PRIMARY KEY NOT NULL,
                                            contact_id INTEGER NOT NULL,
											start_date TEXT,
											type TEXT)`

// CreateTableContactsPhoto creates the table for storing a contact's photo
const CreateTableContactsPhoto = `
CREATE TABLE IF NOT EXISTS contacts_photo (contact_id INTEGER PRIMARY KEY NOT NULL,
                                           photo BLOB NOT NULL)`

// NewSQLiteDB returns a NewtonDB instance that is backed by an SQLiteDB stored
// in a file.
func NewSQLiteDB(dbPath string) (NewtonDB, error) {
	if dbPath == "" {
		return nil, errors.New("dbPath is empty")
	}

	sdb := &SQLiteNewtonDB{}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	sdb.db = sqlx.NewDb(db, "sqlite3")

	err = updateSQLiteDBVersion(sdb)
	if err != nil {
		return nil, err
	}

	return sdb, nil
}

func updateSQLiteDBVersion(sdb *SQLiteNewtonDB) error {
	// make sure the version table exists
	_, err := sdb.db.Exec(CreateTableDatabaseVersion)
	if err != nil {
		return err
	}

	var version int
	err = sdb.db.QueryRow("SELECT version FROM database_version LIMIT 1").Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			// no problem, this is just our first run
			_, err = sdb.db.Exec("INSERT INTO database_version (version) VALUES (0)")
			if err != nil {
				return err
			}
			version = 0
		} else {
			return fmt.Errorf("unable to check sqlitedb version - %v", err)
		}
	}

	switch version {
	case 0:
		err = migrateSQLiteDBFrom0To1(sdb)
	case 1:
	}

	if err != nil {
		return fmt.Errorf("error migrating mariadb schema - %v", err)
	}

	return nil
}

// SQLiteNewtonDB is an SQLite backed implementation of a NewtonDB
type SQLiteNewtonDB struct {
	db *sqlx.DB
}

func migrateSQLiteDBFrom0To1(sdb *SQLiteNewtonDB) error {
	tx, err := sdb.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// create all our tables
	creator := errExecer{tx: tx}
	creator.exec(CreateTableBookmarks)
	creator.exec(CreateTableUsers)
	creator.exec(CreateTableSessions)
	creator.exec(CreateTableContacts)
	creator.exec(CreateTableContactsName)
	creator.exec(CreateTableContactsEmails)
	creator.exec(CreateTableContactsPhones)
	creator.exec(CreateTableContactsIMAccounts)
	creator.exec(CreateTableContactsOrganization)
	creator.exec(CreateTableContactsRelations)
	creator.exec(CreateTableContactsPostalAddresses)
	creator.exec(CreateTableContactsWebsites)
	creator.exec(CreateTableContactsEvents)
	creator.exec(CreateTableContactsPhoto)

	if creator.err != nil {
		return creator.err
	}

	// update the version
	_, err = tx.Exec("UPDATE database_version SET version=1")
	if err != nil {
		return err
	}
	err = tx.Commit()

	return err
}

// Bookmark ...
func (sdb *SQLiteNewtonDB) Bookmark(bookmarkID, ownerID int64) (*Bookmark, error) {
	const selectSQL = `SELECT id, url, title, owner_id FROM bookmarks WHERE id=? AND owner_id=?`
	bookmark := &Bookmark{}
	err := sdb.db.QueryRowx(selectSQL, bookmarkID, ownerID).StructScan(bookmark)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return bookmark, nil
}

// BookmarkExists ...
func (sdb *SQLiteNewtonDB) BookmarkExists(id int64) (bool, error) {
	const existsSQL = `SELECT id FROM bookmarks WHERE id=?`
	var foundID int64
	err := sdb.db.QueryRowx(existsSQL, id).Scan(&foundID)
	switch err {
	case nil:
		return true, nil
	case sql.ErrNoRows:
		return false, nil
	default:
		return false, err
	}
}

// Bookmarks ...
func (sdb *SQLiteNewtonDB) Bookmarks(ownerID int64, pageSize int, page int) ([]*Bookmark, error) {
	builder := squirrel.Select("id, url, title, owner_id").From("bookmarks")
	builder = builder.Where(squirrel.Eq{"owner_id": ownerID})
	if pageSize > 0 {
		builder = builder.Limit(uint64(pageSize))
	}
	if page > 0 {
		builder = builder.Offset(uint64(page * pageSize))
	}
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := sdb.db.Queryx(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookmarks := make([]*Bookmark, 0, 0)
	for rows.Next() {
		b := &Bookmark{}
		rows.StructScan(b)
		bookmarks = append(bookmarks, b)
	}

	return bookmarks, nil
}

// CreateBookmark ...
func (sdb *SQLiteNewtonDB) CreateBookmark(bookmark *Bookmark) (int64, error) {
	const insertSQL = `INSERT INTO bookmarks (url, title, owner_id) VALUES (:url, :title, :owner_id)`
	result, err := sqlx.NamedExec(sdb.db, insertSQL, bookmark)
	if err != nil {
		return -1, err
	}

	return result.LastInsertId()
}

// DeleteBookmark ...
func (sdb *SQLiteNewtonDB) DeleteBookmark(bookmarkID, ownerID int64) error {
	const deleteSQL = `DELETE FROM bookmarks WHERE id=? AND owner_id=?`
	_, err := sdb.db.Exec(deleteSQL, bookmarkID, ownerID)
	return err
}

// EditBookmark ...
func (sdb *SQLiteNewtonDB) EditBookmark(bookmark *Bookmark) error {
	const editSQL = `UPDATE bookmarks SET url=:url, title=:title WHERE id=:id`
	_, err := sdb.db.NamedExec(editSQL, bookmark)
	return err
}

// User ...
func (sdb *SQLiteNewtonDB) User(id int64) (*User, error) {
	const selectSQL = `SELECT id, username, full_name, password FROM users WHERE id=?`
	user := &User{}
	err := sdb.db.QueryRowx(selectSQL, id).StructScan(user)
	switch err {
	case nil:
		return user, nil
	case sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}
}

// UserExists ...
func (sdb *SQLiteNewtonDB) UserExists(id int64) (bool, error) {
	const existsSQL = `SELECT id FROM users WHERE id=?`
	var foundID int64
	err := sdb.db.QueryRowx(existsSQL, id).Scan(&foundID)
	switch err {
	case nil:
		return true, nil
	case sql.ErrNoRows:
		return false, nil
	default:
		return false, err
	}
}

// UserByUsername retrieves a User object by its username
func (sdb *SQLiteNewtonDB) UserByUsername(username string) (*User, error) {
	const selectSQL = `SELECT id, username, full_name, password FROM users WHERE username=?`
	user := &User{}
	err := sdb.db.QueryRowx(selectSQL, username).StructScan(user)
	switch err {
	case nil:
		return user, nil
	case sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}
}

// CreateUser ...
func (sdb *SQLiteNewtonDB) CreateUser(user *User) (int64, error) {
	const insertSQL = `INSERT INTO users (username, full_name, password) VALUES (:username, :full_name, :password)`
	result, err := sqlx.NamedExec(sdb.db, insertSQL, user)
	if err != nil {
		return -1, err
	}

	return result.LastInsertId()
}

// EditUser ...
func (sdb *SQLiteNewtonDB) EditUser(user *User) error {
	const editSQL = `UPDATE users SET username=?, full_name=?, password=?`
	_, err := sdb.db.Exec(editSQL, user.Username, user.FullName, user.Password)
	return err
}

// CreateSession writes a session object to disk and returns the id of the new record
func (sdb *SQLiteNewtonDB) CreateSession(session *Session) (int64, error) {
	const insertSQL = `INSERT INTO sessions (access_token, user_id, creation_date) VALUES (:access_token, :user_id, :creation_date)`
	result, err := sqlx.NamedExec(sdb.db, insertSQL, session)
	if err != nil {
		return -1, err
	}

	return result.LastInsertId()
}

// SessionByAccessToken gets a session from it's access token
func (sdb *SQLiteNewtonDB) SessionByAccessToken(token string) (*Session, error) {
	const selectSQL = `SELECT id, access_token, user_id, creation_date FROM sessions WHERE access_token=?`
	session := &Session{}
	err := sdb.db.QueryRowx(selectSQL, token).StructScan(session)
	switch err {
	case nil:
		return session, nil
	case sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}
}

// CreateContact persists a contact
func (sdb *SQLiteNewtonDB) CreateContact(contact *Contact) (int64, error) {
	const insertSQL = `INSERT INTO contacts (nickname, note, owner_id) VALUES (?, ?, ?)`

	tx, err := sdb.db.Beginx()
	if err != nil {
		return -1, NewtonErr(err)
	}
	defer tx.Rollback()

	result, err := tx.Exec(insertSQL, contact.Nickname, contact.Note, contact.OwnerID)
	if err != nil {
		return -1, NewtonErr(err)
	}
	contactID, err := result.LastInsertId()
	if err != nil {
		return -1, NewtonErr(err)
	}

	// store the name
	const insertNameSQL = `
INSERT INTO contacts_name
	(contact_id, display_name, prefix, given_name, middle_name, family_name, suffix, phonetic_given_name, phonetic_middle_name, phonetic_family_name)
VALUES
	(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	name := contact.Name
	_, err = tx.Exec(insertNameSQL, contactID, name.DisplayName, name.Prefix, name.GivenName, name.MiddleName, name.FamilyName, name.Suffix, name.PhoneticGivenName, name.PhoneticMiddleName, name.PhoneticFamilyName)
	if err != nil {
		return -1, NewtonErr(err)
	}

	// populate any emails in there
	const insertEmailSQL = `INSERT INTO contacts_emails (contact_id, address, type, label) VALUES (?, ?, ?, ?)`
	for _, email := range contact.Emails {
		_, err = tx.Exec(insertEmailSQL, contactID, email.Address, email.Type, email.Label)
		if err != nil {
			return -1, NewtonErr(err)
		}
	}

	// add the phone numbers
	const insertPhoneSQL = `INSERT INTO contacts_phones (contact_id, number, type, label) VALUES (?, ?, ?, ?)`
	for _, phone := range contact.Phones {
		_, err = tx.Exec(insertPhoneSQL, contactID, phone.Number, phone.Type, phone.Label)
		if err != nil {
			return -1, NewtonErr(err)
		}
	}

	// add the IMs
	const insertIMAccountsSQL = `INSERT INTO contacts_im_accounts (contact_id, handle, type, label, protocol, custom_protocol) VALUES (?, ?, ?, ?, ?, ?)`
	for _, account := range contact.IMAccounts {
		_, err = tx.Exec(insertIMAccountsSQL, contactID, account.Handle, account.Type, account.Label, account.Protocol, account.CustomProtocol)
		if err != nil {
			return -1, NewtonErr(err)
		}
	}

	// add the Organization details
	if contact.Org != nil {
		const insertOrgSQL = `INSERT INTO contacts_organization (contact_id, company, title) VALUES (?, ?, ?)`
		_, err = tx.Exec(insertOrgSQL,
			contactID,
			contact.Org.Company,
			contact.Org.Title)
		if err != nil {
			return -1, NewtonErr(err)
		}
	}

	// add the contact's relations
	const insertRelationSQL = `INSERT INTO contacts_relations (contact_id, name, type) VALUES (?, ?,?)`
	for _, relation := range contact.Relations {
		_, err = tx.Exec(insertRelationSQL, contactID, relation.Name, relation.Type)
		if err != nil {
			return -1, NewtonErr(err)
		}
	}

	// add postal addresses
	const insertAddressSQL = `INSERT INTO contacts_postal_addresses (contact_id, street, po_box, neighborhood, city, region, post_code, country, type) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	for _, address := range contact.PostalAddresses {
		_, err = tx.Exec(insertAddressSQL,
			contactID,
			address.Street,
			address.POBox,
			address.Neighborhood,
			address.City,
			address.Region,
			address.PostCode,
			address.Country,
			address.Type)
		if err != nil {
			return -1, NewtonErr(err)
		}
	}

	// add websites
	const insertWebsiteSQL = `INSERT INTO contacts_websites (contact_id, address) VALUES (?, ?)`
	for _, site := range contact.Websites {
		_, err = tx.Exec(insertWebsiteSQL, contactID, site)
		if err != nil {
			return -1, NewtonErr(err)
		}
	}

	// add events
	const insertEventSQL = `INSERT INTO contacts_events (contact_id, start_date, type) VALUES (?, ?, ?)`
	for _, event := range contact.Events {
		_, err = tx.Exec(insertEventSQL, contactID, event.StartDate, event.Type)
		if err != nil {
			return -1, NewtonErr(err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return -1, NewtonErr(err)
	}

	return contactID, nil
}

// ContactExists ...
func (sdb *SQLiteNewtonDB) ContactExists(id int64) (bool, error) {
	const existsSQL = `SELECT id FROM contacts WHERE id=?`
	var foundID int64
	err := sdb.db.QueryRowx(existsSQL, id).Scan(&foundID)
	switch err {
	case nil:
		return true, nil
	case sql.ErrNoRows:
		return false, nil
	default:
		return false, err
	}
}

// Contact ...
func (sdb *SQLiteNewtonDB) Contact(contactID, ownerID int64) (*Contact, error) {
	const contactSQL = `SELECT * FROM contacts WHERE id=? AND owner_id=?`
	contact := &Contact{}
	err := sdb.db.Get(contact, contactSQL, contactID, ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, NewtonErr(err)
	}

	// get the name
	const nameSQL = `
	SELECT display_name,
           prefix,
		   given_name,
		   middle_name,
		   family_name,
		   suffix,
		   phonetic_given_name,
		   phonetic_middle_name,
		   phonetic_family_name
	FROM contacts_name
	WHERE contact_id=?`
	contact.Name = &StructuredName{}
	err = sdb.db.Get(contact.Name, nameSQL, contactID)
	if err != nil {
		return nil, NewtonErr(err)
	}

	// get the emails
	const emailsSQL = `
	SELECT address, type, label
	FROM contacts_emails
	WHERE contact_id=?`
	rows, err := sdb.db.Queryx(emailsSQL, contact.ID)
	if err != nil {
		return nil, NewtonErr(err)
	}
	for rows.Next() {
		email := &Email{}
		err = rows.StructScan(email)
		if err != nil {
			return nil, NewtonErr(err)
		}
		contact.Emails = append(contact.Emails, email)
	}

	// get the phone numbers
	const phonesSQL = `
	SELECT number, type
	FROM contacts_phones
	WHERE contact_id=?`
	rows, err = sdb.db.Queryx(phonesSQL, contact.ID)
	if err != nil {
		return nil, NewtonErr(err)
	}
	for rows.Next() {
		phone := &Phone{}
		err = rows.StructScan(phone)
		if err != nil {
			return nil, NewtonErr(err)
		}
		contact.Phones = append(contact.Phones, phone)
	}

	// get the IM accounts
	const imAccountsSQL = `
	SELECT handle, type, label, protocol, custom_protocol
	FROM contacts_im_accounts
	WHERE contact_id=?`
	rows, err = sdb.db.Queryx(imAccountsSQL, contact.ID)
	if err != nil {
		return nil, NewtonErr(err)
	}
	for rows.Next() {
		account := &IMAccount{}
		err = rows.StructScan(account)
		if err != nil {
			return nil, NewtonErr(err)
		}
		contact.IMAccounts = append(contact.IMAccounts, account)
	}

	// retrieve any organization details
	const orgSQL = `
	SELECT company, title
	FROM contacts_organization
	WHERE contact_id=?`
	org := &Organization{}
	err = sdb.db.QueryRowx(orgSQL, contact.ID).StructScan(org)
	switch err {
	case nil:
		contact.Org = org
		fallthrough
	case sql.ErrNoRows:
	default:
		return nil, NewtonErr(err)
	}

	// retrieve the relations
	const relationsSQL = `SELECT name, type FROM contacts_relations WHERE contact_id=?`
	rows, err = sdb.db.Queryx(relationsSQL, contact.ID)
	if err != nil {
		return nil, NewtonErr(err)
	}
	for rows.Next() {
		relation := &Relation{}
		err = rows.StructScan(relation)
		if err != nil {
			return nil, NewtonErr(err)
		}
		contact.Relations = append(contact.Relations, relation)
	}

	// retrieve postal addresses
	const postalsSQL = `SELECT street, po_box, neighborhood, city, region, post_code, country, type FROM contacts_postal_addresses WHERE contact_id=?`
	rows, err = sdb.db.Queryx(postalsSQL, contact.ID)
	if err != nil {
		return nil, NewtonErr(err)
	}
	for rows.Next() {
		postal := &PostalAddress{}
		err = rows.StructScan(postal)
		if err != nil {
			return nil, NewtonErr(err)
		}
		contact.PostalAddresses = append(contact.PostalAddresses, postal)
	}

	// websites
	const sitesSQL = `SELECT address FROM contacts_websites WHERE contact_id=?`
	rows, err = sdb.db.Queryx(sitesSQL, contact.ID)
	if err != nil {
		return nil, NewtonErr(err)
	}
	for rows.Next() {
		var site string
		err = rows.Scan(&site)
		if err != nil {
			return nil, NewtonErr(err)
		}
		contact.Websites = append(contact.Websites, site)
	}

	// events
	const eventsSQL = `SELECT start_date, type FROM contacts_events WHERE contact_id=?`
	rows, err = sdb.db.Queryx(eventsSQL, contact.ID)
	if err != nil {
		return nil, NewtonErr(err)
	}
	for rows.Next() {
		event := &Event{}
		err = rows.StructScan(event)
		if err != nil {
			return nil, NewtonErr(err)
		}
		contact.Events = append(contact.Events, event)
	}

	return contact, nil
}

// Contacts ...
func (sdb *SQLiteNewtonDB) Contacts(ownerID int64) ([]*Contact, error) {
	// get the ids of all our contacts, then retrieve them using Contact()
	const contactIDsSQL = `SELECT id FROM contacts WHERE owner_id=?`
	rows, err := sdb.db.Queryx(contactIDsSQL, ownerID)
	if err != nil {
		return nil, err
	}

	var contactID int64
	contacts := make([]*Contact, 0, 0)
	for rows.Next() {
		err = rows.Scan(&contactID)
		if err != nil {
			return nil, err
		}
		c, err := sdb.Contact(contactID, ownerID)
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, c)
	}

	return contacts, nil
}

// DeleteContact ...
func (sdb *SQLiteNewtonDB) DeleteContact(contactID, ownerID int64) error {
	tx, err := sdb.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// make sure the contact has the same owner id
	const selectSQL = `SELECT id FROM contacts WHERE id=? AND owner_id=?`
	var foundID int64
	err = tx.QueryRowx(selectSQL, contactID, ownerID).Scan(&foundID)
	if err != nil {
		return err
	}

	deleter := errExecer{tx: tx}
	deleter.exec("DELETE FROM contacts WHERE id=?", contactID)
	deleter.exec("DELETE FROM contacts_name WHERE contact_id=?", contactID)
	deleter.exec("DELETE FROM contacts_emails WHERE contact_id=?", contactID)
	deleter.exec("DELETE FROM contacts_phones WHERE contact_id=?", contactID)
	deleter.exec("DELETE FROM contacts_im_accounts WHERE contact_id=?", contactID)
	deleter.exec("DELETE FROM contacts_organization WHERE contact_id=?", contactID)
	deleter.exec("DELETE FROM contacts_relations WHERE contact_id=?", contactID)
	deleter.exec("DELETE FROM contacts_postal_addresses WHERE contact_id=?", contactID)
	deleter.exec("DELETE FROM contacts_websites WHERE contact_id=?", contactID)
	deleter.exec("DELETE FROM contacts_events WHERE contact_id=?", contactID)
	deleter.exec("DELETE FROM contacts_photo WHERE contact_id=?", contactID)
	if deleter.err != nil {
		return err
	}
	err = tx.Commit()

	return err
}

// SetContactPhoto ...
func (sdb *SQLiteNewtonDB) SetContactPhoto(contactID int64, photo []byte) error {
	// make sure this contact exists
	exists, err := sdb.ContactExists(contactID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("contact %d doesn't exist", contactID)
	}

	if photo != nil {
		const insertSQL = `INSERT OR REPLACE INTO contacts_photo (contact_id, photo) VALUES (?, ?)`
		_, err = sdb.db.Exec(insertSQL, contactID, photo)
		return err
	}

	// nil photo, so this is a deletion
	_, err = sdb.db.Exec("DELETE FROM contacts_photo WHERE contact_id=?", contactID)
	return err
}

// ContactPhoto ...
func (sdb *SQLiteNewtonDB) ContactPhoto(contactID int64) ([]byte, error) {
	const selectSQL = `SELECT photo FROM contacts_photo WHERE contact_id=?`
	var photo []byte
	err := sdb.db.QueryRowx(selectSQL, contactID).Scan(&photo)
	switch err {
	case nil:
		return photo, nil
	case sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}
}

// ContactOwner ...
func (sdb *SQLiteNewtonDB) ContactOwner(contactID int64) (int64, error) {
	var ownerID int64
	err := sdb.db.QueryRow("SELECT owner_id FROM contacts WHERE id=?", contactID).Scan(&ownerID)
	return ownerID, err
}
