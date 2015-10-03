package main

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// CreateTableDatabaseVersion is the statement to create a table that tracks the current schema version
const CreateTableDatabaseVersion = `CREATE TABLE IF NOT EXISTS database_version (id INT NOT NULL AUTO_INCREMENT, version INT NOT NULL DEFAULT 0, PRIMARY KEY (id))`

// CreateTableBookmarks is the statement to create the bookmarks table
const CreateTableBookmarks = `CREATE TABLE IF NOT EXISTS bookmarks (id INT NOT NULL AUTO_INCREMENT, url VARCHAR(2048) NOT NULL, title VARCHAR(1024), owner_id INT NOT NULL, PRIMARY KEY (id))`

// CreateTableUsers is the statement to create the users table
const CreateTableUsers = `CREATE TABLE IF NOT EXISTS users (id INT NOT NULL AUTO_INCREMENT, username VARCHAR(64) NOT NULL, full_name VARCHAR(128) NOT NULL, password VARCHAR(96) NOT NULL, PRIMARY KEY (id))`

// CreateTableSessions is the statement to create the sessions table
const CreateTableSessions = `CREATE TABLE IF NOT EXISTS sessions (id INT NOT NULL AUTO_INCREMENT, access_token CHAR(32) NOT NULL, user_id INT NOT NULL, creation_date DATETIME NOT NULL, PRIMARY KEY (id))`

// CreateTableContacts is the statement to create the contacts table
const CreateTableContacts = `CREATE TABLE IF NOT EXISTS contacts (id INT NOT NULL AUTO_INCREMENT, name VARCHAR(256), owner_id INT NOT NULL, PRIMARY KEY (id))`

// CreateTableContactsEmails creates the table for storing a contact's emails
const CreateTableContactsEmails = `CREATE TABLE IF NOT EXISTS contacts_emails (id INT NOT NULL AUTO_INCREMENT, contact_id INT NOT NULL, address VARCHAR(320), type VARCHAR(32), PRIMARY KEY (id))`

// CreateTableContactsPhones creates the table for storing a contact's phone numbers
const CreateTableContactsPhones = `CREATE TABLE IF NOT EXISTS contacts_phones (id INT NOT NULL AUTO_INCREMENT, contact_id INT NOT NULL, number VARCHAR(128), type VARCHAR(32), PRIMARY KEY (id))`

// CreateTableContactsIMHandles creates the table for storing a contact's IM handles
const CreateTableContactsIMHandles = `CREATE TABLE IF NOT EXISTS contacts_im_handles (id INT NOT NULL AUTO_INCREMENT, contact_id INT NOT NULL, identifier VARCHAR(128), protocol VARCHAR(32), type VARCHAR(32), PRIMARY KEY (id))`

// CreateTableContactsOrganization creates the table for storing a contact's organization/association details
const CreateTableContactsOrganization = `CREATE TABLE IF NOT EXISTS contacts_organization (contact_id INT NOT NULL,
	company VARCHAR(128),
	type VARCHAR(32),
	title VARCHAR(64),
	department VARCHAR(64),
	job_description VARCHAR(64),
	symbol VARCHAR(16),
	phonetic_name VARCHAR(64),
	office_location VARCHAR(64), PRIMARY KEY (contact_id))`

// CreateTableContactsRelations creates the table for storing a contact's relations (spouse, children, etc.)
const CreateTableContactsRelations = `CREATE TABLE IF NOT EXISTS contacts_relations (id INT NOT NULL AUTO_INCREMENT, contact_id INT NOT NULL, name VARCHAR(128), type VARCHAR(32), PRIMARY KEY (id))`

// CreateTableContactsPostalAddresses creates the table for storing a contact's postal addresses
const CreateTableContactsPostalAddresses = `CREATE TABLE IF NOT EXISTS contacts_postal_addresses (id INT NOT NULL AUTO_INCREMENT, contact_id INT NOT NULL, street VARCHAR(256), po_box VARCHAR(16), neighborhood VARCHAR(128), city VARCHAR(128), region VARCHAR(128), post_code VARCHAR(16), country VARCHAR(96), type VARCHAR(32), PRIMARY KEY(id))`

// CreateTableContactsWebsites creates the table for storing a contact's websites
const CreateTableContactsWebsites = `CREATE TABLE IF NOT EXISTS contacts_websites (id INT NOT NULL AUTO_INCREMENT, contact_id INT NOT NULL, address VARCHAR(2048), type VARCHAR(32), PRIMARY KEY (id))`

// CreateTableContactsEvents creates the table for storing a contact's events
const CreateTableContactsEvents = `CREATE TABLE IF NOT EXISTS contacts_events (id INT NOT NULL AUTO_INCREMENT, contact_id INT NOT NULL, start_date VARCHAR(48), type VARCHAR(32), PRIMARY KEY (id))`

// NewMariaDB returns a NewtonDB instance that is backed by a MariaDB instance described
// in the dsn.
func NewMariaDB(dsn string) (NewtonDB, error) {
	if dsn == "" {
		return nil, errors.New("dsn is empty")
	}

	mdb := &MariaNewtonDB{}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	mdb.db = sqlx.NewDb(db, "mysql")

	err = mdb.db.Ping()
	if err != nil {
		return nil, fmt.Errorf("mariadb ping failed - %v", err)
	}

	// make sure the database is up to date
	err = updateMariaDBVersion(mdb)
	if err != nil {
		return nil, err
	}

	return mdb, nil
}

func updateMariaDBVersion(mdb *MariaNewtonDB) error {
	// make sure the version table exists
	_, err := mdb.db.Exec(CreateTableDatabaseVersion)
	if err != nil {
		return err
	}

	var version int
	err = mdb.db.QueryRow("SELECT version FROM database_version LIMIT 1").Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			// no problem, this is just our first run
			_, err = mdb.db.Exec("INSERT INTO database_version (version) VALUES (0)")
			if err != nil {
				return err
			}
			version = 0
		} else {
			return fmt.Errorf("unable to check mariadb version - %v", err)
		}
	}

	switch version {
	case 0:
		err = migrateMariaDBFrom0To1(mdb)
	case 1:
	}

	if err != nil {
		return fmt.Errorf("error migrating mariadb schema - %v", err)
	}

	return nil
}

type errTableCreator struct {
	tx  *sqlx.Tx
	err error
}

func (etc *errTableCreator) exec(query string) {
	if etc.err != nil {
		return
	}

	_, etc.err = etc.tx.Exec(query)
}

func migrateMariaDBFrom0To1(mdb *MariaNewtonDB) error {
	tx, err := mdb.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// create all our tables
	creator := errTableCreator{tx: tx}
	creator.exec(CreateTableBookmarks)
	creator.exec(CreateTableUsers)
	creator.exec(CreateTableSessions)
	creator.exec(CreateTableContacts)
	creator.exec(CreateTableContactsEmails)
	creator.exec(CreateTableContactsPhones)
	creator.exec(CreateTableContactsIMHandles)
	creator.exec(CreateTableContactsOrganization)
	creator.exec(CreateTableContactsRelations)
	creator.exec(CreateTableContactsPostalAddresses)
	creator.exec(CreateTableContactsWebsites)
	creator.exec(CreateTableContactsEvents)

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

// MariaNewtonDB is...
type MariaNewtonDB struct {
	db *sqlx.DB
}

// CreateBookmark creates a bookmark
func (mdb *MariaNewtonDB) CreateBookmark(bookmark *Bookmark) (int64, error) {
	const insertSQL = `INSERT INTO bookmarks (url, title, owner_id) VALUES (:url, :title, :owner_id)`
	result, err := sqlx.NamedExec(mdb.db, insertSQL, bookmark)
	if err != nil {
		return -1, err
	}

	return result.LastInsertId()
}

// Bookmark retrieves a bookmark by its id
func (mdb *MariaNewtonDB) Bookmark(id int64) (*Bookmark, error) {
	const selectSQL = `SELECT id, url, title, owner_id FROM bookmarks WHERE id=?`
	bookmark := &Bookmark{}
	err := mdb.db.QueryRowx(selectSQL, id).StructScan(bookmark)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return bookmark, nil
}

// BookmarkExists returns true if the bookmark exists
func (mdb *MariaNewtonDB) BookmarkExists(id int64) (bool, error) {
	const existsSQL = `SELECT id FROM bookmarks WHERE id=?`
	var foundID int64
	err := mdb.db.QueryRowx(existsSQL, id).Scan(&foundID)
	switch err {
	case nil:
		return true, nil
	case sql.ErrNoRows:
		return false, nil
	default:
		return false, err
	}
}

// DeleteBookmark deletes a bookmark with the specified id
func (mdb *MariaNewtonDB) DeleteBookmark(id int64) error {
	const deleteSQL = `DELETE FROM bookmarks WHERE id=?`
	_, err := mdb.db.Exec(deleteSQL, id)
	return err
}

// Bookmarks retrieves a list of bookmarks according to the specified arguments
func (mdb *MariaNewtonDB) Bookmarks(userID int64, pageSize int, page int) ([]*Bookmark, error) {
	builder := squirrel.Select("id, url, title, owner_id").From("bookmarks")
	builder = builder.Where(squirrel.Eq{"owner_id": userID})
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
	rows, err := mdb.db.Queryx(query, args...)
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

// EditBookmark edits an existing bookmark
func (mdb *MariaNewtonDB) EditBookmark(bookmark *Bookmark) error {
	const editSQL = `UPDATE bookmarks SET url=:url, title=:title WHERE id=:id`
	_, err := mdb.db.NamedExec(editSQL, bookmark)
	return err
}

// UserByUsername retrieves a User object by its username
func (mdb *MariaNewtonDB) UserByUsername(username string) (*User, error) {
	const selectSQL = `SELECT id, username, full_name, password FROM users WHERE username=?`
	user := &User{}
	err := mdb.db.QueryRowx(selectSQL, username).StructScan(user)
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
func (mdb *MariaNewtonDB) CreateUser(user *User) (int64, error) {
	const insertSQL = `INSERT INTO users (username, full_name, password) VALUES (:username, :full_name, :password)`
	result, err := sqlx.NamedExec(mdb.db, insertSQL, user)
	if err != nil {
		return -1, err
	}

	return result.LastInsertId()
}

// EditUser ...
func (mdb *MariaNewtonDB) EditUser(user *User) error {
	const editSQL = `UPDATE users SET username=?, full_name=?, password=?`
	_, err := mdb.db.Exec(editSQL, user.Username, user.FullName, user.Password)
	return err
}

// UserExists ...
func (mdb *MariaNewtonDB) UserExists(id int64) (bool, error) {
	const existsSQL = `SELECT id FROM users WHERE id=?`
	var foundID int64
	err := mdb.db.QueryRowx(existsSQL, id).Scan(&foundID)
	switch err {
	case nil:
		return true, nil
	case sql.ErrNoRows:
		return false, nil
	default:
		return false, err
	}
}

// User ...
func (mdb *MariaNewtonDB) User(id int64) (*User, error) {
	const selectSQL = `SELECT id, username, full_name, password FROM users WHERE id=?`
	user := &User{}
	err := mdb.db.QueryRowx(selectSQL, id).StructScan(user)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

// CreateSession writes a session object to disk and returns the id of the new record
func (mdb *MariaNewtonDB) CreateSession(session *Session) (int64, error) {
	const insertSQL = `INSERT INTO sessions (access_token, user_id, creation_date) VALUES (:access_token, :user_id, :creation_date)`
	result, err := sqlx.NamedExec(mdb.db, insertSQL, session)
	if err != nil {
		return -1, err
	}

	return result.LastInsertId()
}

// SessionByAccessToken gets a session from it's access token
func (mdb *MariaNewtonDB) SessionByAccessToken(token string) (*Session, error) {
	const selectSQL = `SELECT id, access_token, user_id, creation_date FROM sessions WHERE access_token=?`
	session := &Session{}
	err := mdb.db.QueryRowx(selectSQL, token).StructScan(session)
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
func (mdb *MariaNewtonDB) CreateContact(contact *Contact) (int64, error) {
	const insertSQL = `INSERT INTO contacts (name, owner_id) VALUES (:name, :owner_id)`

	tx, err := mdb.db.Beginx()
	if err != nil {
		return -1, err
	}
	defer tx.Rollback()

	result, err := sqlx.NamedExec(tx, insertSQL, contact)
	if err != nil {
		return -1, err
	}
	contactID, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}

	// populate any emails in there
	const insertEmailSQL = `INSERT INTO contacts_emails (contact_id, address, type) VALUES (?, ?, ?)`
	for _, email := range contact.Emails {
		_, err = tx.Exec(insertEmailSQL, contactID, email.Address, email.Type)
		if err != nil {
			return -1, err
		}
	}

	// add the phone numbers
	const insertPhoneSQL = `INSERT INTO contacts_phones (contact_id, number, type) VALUES (?, ?, ?)`
	for _, phone := range contact.Phones {
		_, err = tx.Exec(insertPhoneSQL, contactID, phone.Number, phone.Type)
		if err != nil {
			return -1, err
		}
	}

	// add the IM handles
	const insertIMHandleSQL = `INSERT INTO contacts_im_handles (contact_id, identifier, protocol, type) VALUES (?, ?, ?, ?)`
	for _, imHandle := range contact.IMHandles {
		_, err = tx.Exec(insertIMHandleSQL, contactID, imHandle.Identifier, imHandle.Protocol, imHandle.Type)
		if err != nil {
			return -1, err
		}
	}

	// add the Organization details
	if contact.Org != nil {
		const insertOrgSQL = `INSERT INTO contacts_organization (contact_id, company, type, title, department, job_description, symbol, phonetic_name, office_location) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
		_, err = tx.Exec(insertOrgSQL,
			contactID,
			contact.Org.Company,
			contact.Org.Type,
			contact.Org.Title,
			contact.Org.Department,
			contact.Org.JobDescription,
			contact.Org.Symbol,
			contact.Org.PhoneticName,
			contact.Org.OfficeLocation)
		if err != nil {
			return -1, err
		}
	}

	// add the contact's relations
	const insertRelationSQL = `INSERT INTO contacts_relations (contact_id, name, type) VALUES (?, ?,?)`
	for _, relation := range contact.Relations {
		_, err = tx.Exec(insertRelationSQL, contactID, relation.Name, relation.Type)
		if err != nil {
			return -1, err
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
			return -1, err
		}
	}

	// add websites
	const insertWebsiteSQL = `INSERT INTO contacts_websites (contact_id, address, type) VALUES (?, ?, ?)`
	for _, site := range contact.Websites {
		_, err = tx.Exec(insertWebsiteSQL, contactID, site.Address, site.Type)
		if err != nil {
			return -1, err
		}
	}

	// add events
	const insertEventSQL = `INSERT INTO contacts_events (contact_id, start_date, type) VALUES (?, ?, ?)`
	for _, event := range contact.Events {
		_, err = tx.Exec(insertEventSQL, contactID, event.StartDate, event.Type)
		if err != nil {
			return -1, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return -1, err
	}

	return contactID, nil
}
