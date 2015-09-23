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
const CreateTableSessions = `CREATE TABLE IF NOT EXISTS sessions (id INT NOT NULL AUTO_INCRMENT, access_token CHAR(32) NOT NULL, user_id INT NOT NULL, creation_date DATETIME NOT NULL, PRIMARY KEY (id))`

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

func migrateMariaDBFrom0To1(mdb *MariaNewtonDB) error {
	tx, err := mdb.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// create all our tables
	_, err = tx.Exec(CreateTableBookmarks)
	if err != nil {
		return err
	}
	_, err = tx.Exec(CreateTableUsers)
	if err != nil {
		return err
	}
	_, err = tx.Exec(CreateTableSessions)
	if err != nil {
		return err
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
	const insertSQL = `INSERT INTO bookmarks (url, title, user_id) VALUES (:url, :title, :user_id)`
	query, args, err := sqlx.Named(insertSQL, bookmark)
	if err != nil {
		return -1, err
	}

	result, err := mdb.db.Exec(query, args...)
	if err != nil {
		return -1, err
	}

	return result.LastInsertId()
}

// Bookmark retrieves a bookmark by its id
func (mdb *MariaNewtonDB) Bookmark(id int64) (*Bookmark, error) {
	const selectSQL = `SELECT id, url, title, user_id FROM bookmarks WHERE id=?`
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
func (mdb *MariaNewtonDB) Bookmarks(filters map[string]interface{}, pageSize int, page int) ([]*Bookmark, error) {
	builder := squirrel.Select("bookmarks.id, bookmarks.url, bookmarks.title, bookmarks.user_id").From("bookmarks")
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
	const editSQL = `UPDATE bookmarks SET url=?, title=? WHERE id=?`
	_, err := mdb.db.Exec(editSQL, bookmark.URL, bookmark.Title, bookmark.ID)
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
