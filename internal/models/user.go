package models

import (
    "database/sql"
    "errors"

    "github.com/google/uuid"
)

type User struct {
    ID uuid.UUID
    Username string
    Email string
    IsDeleted bool
}

type UserModel struct {
    DB *sql.DB
}

func (m *UserModel) Insert(username string, email string) (uuid.UUID, error) {
    id := uuid.New()
    stmt := `INSERT INTO users (Id, Username, Email, IsDeleted)
    VALUES (?, ?, ?, FALSE)`
    _, err := m.DB.Exec(stmt, id, username, email)
    if err != nil {
	return uuid.UUID{}, err
    }
    return id, nil
}

func (m *UserModel) Get(id uuid.UUID) (User, error) {
    stmt := `SELECT Id, Username, Email, IsDeleted FROM users WHERE id = ?`
    row := m.DB.QueryRow(stmt, id)
    var u User
    err := row.Scan(&u.ID, &u.Username, &u.Email, &u.IsDeleted)
    if err != nil {
	if errors.Is(err, sql.ErrNoRows) {
	    return User{}, ErrNoRecord
	}
	return User{}, err
    }
    return u, nil
}

func (m *UserModel) GetAll() ([]User, error) {
    stmt := `SELECT Id, Username, Email, IsDeleted FROM users`
    rows, err := m.DB.Query(stmt)
    var users []User
    for rows.Next() {
	var u User
	err = rows.Scan(&u.ID, &u.Username, &u.Email, &u.IsDeleted)
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
