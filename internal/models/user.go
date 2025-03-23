package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int64
	ShortId        uint64
	Username       string
	Email          string
	Deleted        bool
	IsBanned       bool
	HashedPassword []byte
	Created        time.Time
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(shortId uint64, username string, email string, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	stmt := `INSERT INTO users (ShortId, Username, Email, IsBanned, HashedPassword, Created)
    VALUES (?, ?, ?, FALSE, ?, ?)`
	_, err = m.DB.Exec(stmt, shortId, username, email, hashedPassword, time.Now().UTC())
	if err != nil {
		if sqliteError, ok := err.(sqlite3.Error); ok {
			if sqliteError.ExtendedCode == 2067 && strings.Contains(sqliteError.Error(), "Email") {
				return ErrDuplicateEmail
			}
		}
		return err
	}
	return nil
}

func (m *UserModel) Get(id uint64) (User, error) {
	stmt := `SELECT Id, ShortId, Username, Email, Created FROM users WHERE ShortId = ? AND Deleted IS NULL`
	row := m.DB.QueryRow(stmt, id)
	var u User
	err := row.Scan(&u.ID, &u.ShortId, &u.Username, &u.Email, &u.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNoRecord
		}
		return User{}, err
	}
	return u, nil
}

func (m *UserModel) GetById(id int64) (User, error) {
	stmt := `SELECT Id, ShortId, Username, Email, Created FROM users WHERE Id = ? AND Deleted IS NULL`
	row := m.DB.QueryRow(stmt, id)
	var u User
	err := row.Scan(&u.ID, &u.ShortId, &u.Username, &u.Email, &u.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNoRecord
		}
		return User{}, err
	}
	return u, nil
}

func (m *UserModel) GetAll() ([]User, error) {
	stmt := `SELECT Id, ShortId, Username, Email, Created FROM users WHERE DELETED IS NULL`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	var users []User
	for rows.Next() {
		var u User
		err = rows.Scan(&u.ID, &u.ShortId, &u.Username, &u.Email, &u.Created)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (m *UserModel) Authenticate(email, password string) (int64, error) {
	var id int64
	var hashedPassword []byte

	stmt := `SELECT Id, HashedPassword FROM users WHERE Email = ?`
	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	return id, nil
}

func (m *UserModel) Exists(id int64) (bool, error) {
	var exists bool
	stmt := `SELECT EXISTS(SELECT true FROM users WHERE Id = ? AND DELETED IS NULL)`
	err := m.DB.QueryRow(stmt, id).Scan(&exists)
	return exists, err
}
