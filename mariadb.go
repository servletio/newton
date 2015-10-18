package main

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// DropAllMariaDBTables is just useful when testing
const DropAllMariaDBTables = `
DROP TABLE bookmarks,
           contacts,
		   contacts_name,
		   contacts_emails,
		   contacts_events,
		   ontacts_im_handles,
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
CREATE TABLE IF NOT EXISTS database_version (id INT NOT NULL AUTO_INCREMENT,
                                             version INT NOT NULL DEFAULT 0,
											 PRIMARY KEY (id))`

// CreateTableBookmarks is the statement to create the bookmarks table
const CreateTableBookmarks = `
CREATE TABLE IF NOT EXISTS bookmarks (id INT NOT NULL AUTO_INCREMENT,
                                      url VARCHAR(2048) NOT NULL,
									  title VARCHAR(1024),
									  owner_id INT NOT NULL,
									  PRIMARY KEY (id))`

// CreateTableUsers is the statement to create the users table
const CreateTableUsers = `
CREATE TABLE IF NOT EXISTS users (id INT NOT NULL AUTO_INCREMENT,
                                  username VARCHAR(64) NOT NULL,
								  full_name VARCHAR(128) NOT NULL,
								  password VARCHAR(96) NOT NULL,
								  PRIMARY KEY (id))`

// CreateTableSessions is the statement to create the sessions table
const CreateTableSessions = `
CREATE TABLE IF NOT EXISTS sessions (id INT NOT NULL AUTO_INCREMENT,
                                     access_token CHAR(32) NOT NULL,
									 user_id INT NOT NULL,
									 creation_date DATETIME NOT NULL,
									 PRIMARY KEY (id))`

// CreateTableContacts is the statement to create the contacts table
const CreateTableContacts = `
CREATE TABLE IF NOT EXISTS contacts (id INT NOT NULL AUTO_INCREMENT,
                                     owner_id INT NOT NULL,
									 PRIMARY KEY (id))`

// CreateTableContactsName is the statement to create the table for storing a contact's name
const CreateTableContactsName = `
CREATE TABLE IF NOT EXISTS contacts_name (contact_id INT NOT NULL,
                                          display_name VARCHAR(256),
										  prefix VARCHAR(32),
										  given_name VARCHAR(64),
										  middle_name VARCHAR(64),
										  family_name VARCHAR(64),
										  suffix VARCHAR(32),
										  phonetic_given_name VARCHAR(64),
										  phonetic_middle_name VARCHAR(64),
										  phonetic_family_name VARCHAR(64),
										  PRIMARY KEY (contact_id))`

// CreateTableContactsEmails creates the table for storing a contact's emails
const CreateTableContactsEmails = `
CREATE TABLE IF NOT EXISTS contacts_emails (id INT NOT NULL AUTO_INCREMENT,
                                            contact_id INT NOT NULL,
											address VARCHAR(320),
											type VARCHAR(32),
											PRIMARY KEY (id))`

// CreateTableContactsPhones creates the table for storing a contact's phone numbers
const CreateTableContactsPhones = `
CREATE TABLE IF NOT EXISTS contacts_phones (id INT NOT NULL AUTO_INCREMENT,
                                            contact_id INT NOT NULL,
											number VARCHAR(128),
											type VARCHAR(32),
											PRIMARY KEY (id))`

// CreateTableContactsIMHandles creates the table for storing a contact's IM handles
const CreateTableContactsIMHandles = `
CREATE TABLE IF NOT EXISTS contacts_im_handles (id INT NOT NULL AUTO_INCREMENT,
                                                contact_id INT NOT NULL,
												identifier VARCHAR(128),
												protocol VARCHAR(32),
												type VARCHAR(32),
												PRIMARY KEY (id))`

// CreateTableContactsOrganization creates the table for storing a contact's organization/association details
const CreateTableContactsOrganization = `
CREATE TABLE IF NOT EXISTS contacts_organization (contact_id INT NOT NULL,
                                                  company VARCHAR(128),
												  type VARCHAR(32),
												  title VARCHAR(64),
												  department VARCHAR(64),
												  job_description VARCHAR(64),
												  symbol VARCHAR(16),
												  phonetic_name VARCHAR(64),
												  office_location VARCHAR(64), PRIMARY KEY (contact_id))`

// CreateTableContactsRelations creates the table for storing a contact's relations (spouse, children, etc.)
const CreateTableContactsRelations = `
CREATE TABLE IF NOT EXISTS contacts_relations (id INT NOT NULL AUTO_INCREMENT,
                                               contact_id INT NOT NULL,
											   name VARCHAR(128),
											   type VARCHAR(32),
											   PRIMARY KEY (id))`

// CreateTableContactsPostalAddresses creates the table for storing a contact's postal addresses
const CreateTableContactsPostalAddresses = `
CREATE TABLE IF NOT EXISTS contacts_postal_addresses (id INT NOT NULL AUTO_INCREMENT,
                                                      contact_id INT NOT NULL,
													  street VARCHAR(256),
													  po_box VARCHAR(16),
													  neighborhood VARCHAR(128),
													  city VARCHAR(128),
													  region VARCHAR(128),
													  post_code VARCHAR(16),
													  country VARCHAR(96),
													  type VARCHAR(32),
													  PRIMARY KEY(id))`

// CreateTableContactsWebsites creates the table for storing a contact's websites
const CreateTableContactsWebsites = `
CREATE TABLE IF NOT EXISTS contacts_websites (id INT NOT NULL AUTO_INCREMENT,
                                              contact_id INT NOT NULL,
											  address VARCHAR(2048),
											  type VARCHAR(32),
											  PRIMARY KEY (id))`

// CreateTableContactsEvents creates the table for storing a contact's events
const CreateTableContactsEvents = `
CREATE TABLE IF NOT EXISTS contacts_events (id INT NOT NULL AUTO_INCREMENT,
                                            contact_id INT NOT NULL,
											start_date VARCHAR(48),
											type VARCHAR(32),
											PRIMARY KEY (id))`

// CreateTableContactsPhoto creates the table for storing a contact's photo
const CreateTableContactsPhoto = `
CREATE TABLE IF NOT EXISTS contacts_photo (contact_id INT NOT NULL,
                                           photo MEDIUMBLOB NOT NULL,
										   PRIMARY KEY (contact_id))`

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

type errExecer struct {
	tx  *sqlx.Tx
	err error
}

func (ee *errExecer) exec(query string, args ...interface{}) {
	if ee.err != nil {
		return
	}

	_, ee.err = ee.tx.Exec(query, args...)
}

func migrateMariaDBFrom0To1(mdb *MariaNewtonDB) error {
	tx, err := mdb.db.Beginx()
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
	creator.exec(CreateTableContactsIMHandles)
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
func (mdb *MariaNewtonDB) Bookmark(bookmarkID, ownerID int64) (*Bookmark, error) {
	const selectSQL = `SELECT id, url, title, owner_id FROM bookmarks WHERE id=? AND owner_id=?`
	bookmark := &Bookmark{}
	err := mdb.db.QueryRowx(selectSQL, bookmarkID, ownerID).StructScan(bookmark)
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
func (mdb *MariaNewtonDB) DeleteBookmark(bookmarkID, ownerID int64) error {
	const deleteSQL = `DELETE FROM bookmarks WHERE id=? AND owner_id=?`
	_, err := mdb.db.Exec(deleteSQL, bookmarkID, ownerID)
	return err
}

// Bookmarks retrieves a list of bookmarks according to the specified arguments
func (mdb *MariaNewtonDB) Bookmarks(ownerID int64, pageSize int, page int) ([]*Bookmark, error) {
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
	const insertSQL = `INSERT INTO contacts (owner_id) VALUES (:owner_id)`

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

	// store the name
	const insertNameSQL = `
INSERT INTO contacts_name
	(contact_id, display_name, prefix, given_name, middle_name, family_name, suffix, phonetic_given_name, phonetic_middle_name, phonetic_family_name)
VALUES
	(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	name := contact.Name
	_, err = tx.Exec(insertNameSQL, contactID, name.DisplayName, name.Prefix, name.GivenName, name.MiddleName, name.FamilyName, name.Suffix, name.PhoneticGivenName, name.PhoneticMiddleName, name.PhoneticFamilyName)
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

// ContactExists ...
func (mdb *MariaNewtonDB) ContactExists(id int64) (bool, error) {
	const existsSQL = `SELECT id FROM contacts WHERE id=?`
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

// Contact ...
func (mdb *MariaNewtonDB) Contact(contactID, ownerID int64) (*Contact, error) {
	const contactSQL = `SELECT id, owner_id FROM contacts WHERE id=? AND owner_id=?`
	contact := &Contact{}
	err := mdb.db.Get(contact, contactSQL, contactID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// get the name
	const nameSQL = `SELECT * FROM contacts_name WHERE id=?`
	contact.Name = &StructuredName{}
	err = mdb.db.Get(contact.Name, nameSQL, contactID)
	if err != nil {
		return nil, err
	}

	// get the emails
	const emailsSQL = `SELECT address, type FROM contacts_emails WHERE contact_id=?`
	rows, err := mdb.db.Queryx(emailsSQL, contact.ID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		email := &Email{}
		err = rows.StructScan(email)
		if err != nil {
			return nil, err
		}
		contact.Emails = append(contact.Emails, email)
	}

	// get the phone numbers
	const phonesSQL = `SELECT number, type FROM contacts_phones WHERE contact_id=?`
	rows, err = mdb.db.Queryx(phonesSQL, contact.ID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		phone := &Phone{}
		err = rows.StructScan(phone)
		if err != nil {
			return nil, err
		}
		contact.Phones = append(contact.Phones, phone)
	}

	// get the IM handles
	const imHandlesSQL = `SELECT identifier, protocol, type FROM contacts_im_handles WHERE contact_id=?`
	rows, err = mdb.db.Queryx(imHandlesSQL, contact.ID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		imHandle := &IMHandle{}
		err = rows.StructScan(imHandle)
		if err != nil {
			return nil, err
		}
		contact.IMHandles = append(contact.IMHandles, imHandle)
	}

	// retrieve any organization details
	const orgSQL = `SELECT company, type, title, department, job_description, symbol, phonetic_name, office_location FROM contacts_organization WHERE contact_id=?`
	org := &Organization{}
	err = mdb.db.QueryRowx(orgSQL, contact.ID).StructScan(org)
	switch err {
	case nil:
		contact.Org = org
		fallthrough
	case sql.ErrNoRows:
	default:
		return nil, err
	}

	// retrieve the relations
	const relationsSQL = `SELECT name, type FROM contacts_relations WHERE contact_id=?`
	rows, err = mdb.db.Queryx(relationsSQL, contact.ID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		relation := &Relation{}
		err = rows.StructScan(relation)
		if err != nil {
			return nil, err
		}
		contact.Relations = append(contact.Relations, relation)
	}

	// retrieve postal addresses
	const postalsSQL = `SELECT street, po_box, neighborhood, city, region, post_code, country, type FROM contacts_postal_addresses WHERE contact_id=?`
	rows, err = mdb.db.Queryx(postalsSQL, contact.ID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		postal := &PostalAddress{}
		err = rows.StructScan(postal)
		if err != nil {
			return nil, err
		}
		contact.PostalAddresses = append(contact.PostalAddresses, postal)
	}

	// websites
	const sitesSQL = `SELECT address, type FROM contacts_websites WHERE contact_id=?`
	rows, err = mdb.db.Queryx(sitesSQL, contact.ID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		site := &Website{}
		err = rows.StructScan(site)
		if err != nil {
			return nil, err
		}
		contact.Websites = append(contact.Websites, site)
	}

	// events
	const eventsSQL = `SELECT start_date, type FROM contacts_events WHERE contact_id=?`
	rows, err = mdb.db.Queryx(eventsSQL, contact.ID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		event := &Event{}
		err = rows.StructScan(event)
		if err != nil {
			return nil, err
		}
		contact.Events = append(contact.Events, event)
	}

	return contact, nil
}

// Contacts ...
func (mdb *MariaNewtonDB) Contacts(ownerID int64) ([]*Contact, error) {
	// get the ids of all our contacts, then retrieve them using Contact()
	const contactIDsSQL = `SELECT id FROM contacts WHERE owner_id=?`
	rows, err := mdb.db.Queryx(contactIDsSQL, ownerID)
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
		c, err := mdb.Contact(contactID, ownerID)
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, c)
	}

	return contacts, nil
}

// DeleteContact ...
func (mdb *MariaNewtonDB) DeleteContact(contactID, ownerID int64) error {
	tx, err := mdb.db.Beginx()
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
	deleter.exec("DELETE FROM contacts_im_handles WHERE contact_id=?", contactID)
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
func (mdb *MariaNewtonDB) SetContactPhoto(contactID int64, photo []byte) error {
	// make sure this contact exists
	exists, err := mdb.ContactExists(contactID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("contact %d doesn't exist", contactID)
	}

	if photo != nil {
		const insertSQL = `INSERT INTO contacts_photo (contact_id, photo) VALUES (?, ?) ON DUPLICATE KEY UPDATE photo=?`
		_, err = mdb.db.Exec(insertSQL, contactID, photo, photo)
		return err
	}

	// nil photo, so this is a deletion
	_, err = mdb.db.Exec("DELETE FROM contacts_photo WHERE contact_id=?", contactID)
	return err
}

// ContactPhoto ...
func (mdb *MariaNewtonDB) ContactPhoto(contactID int64) ([]byte, error) {
	const selectSQL = `SELECT photo FROM contacts_photo WHERE contact_id=?`
	var photo []byte
	err := mdb.db.QueryRowx(selectSQL, contactID).Scan(&photo)
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
func (mdb *MariaNewtonDB) ContactOwner(contactID int64) (int64, error) {
	var ownerID int64
	err := mdb.db.QueryRow("SELECT owner_id FROM contacts WHERE id=?", contactID).Scan(&ownerID)
	return ownerID, err
}
