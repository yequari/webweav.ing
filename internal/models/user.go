package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const (
	u_timezone = 1
)

type UserSettings struct {
	LocalTimezone *time.Location
}

type User struct {
	ID             int64
	ShortId        uint64
	Username       string
	Email          string
	Deleted        bool
	IsBanned       bool
	HashedPassword []byte
	Created        time.Time
	Settings       UserSettings
}

type UserModel struct {
	DB       *sql.DB
	Settings map[string]Setting
}

func (m *UserModel) Insert(shortId uint64, username string, email string, password string, settings UserSettings) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	stmt := `INSERT INTO users (ShortId, Username, Email, IsBanned, HashedPassword, Created)
    VALUES (?, ?, ?, FALSE, ?, ?)`
	result, err := m.DB.Exec(stmt, shortId, username, email, hashedPassword, time.Now().UTC())
	if err != nil {
		if sqliteError, ok := err.(sqlite3.Error); ok {
			if sqliteError.ExtendedCode == 2067 && strings.Contains(sqliteError.Error(), "Email") {
				return ErrDuplicateEmail
			}
		}
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	err = m.initializeUserSettings(id, settings)
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
	settings, err := m.GetSettings(u.ID)
	if err != nil {
		return u, err
	}
	u.Settings = settings
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
	settings, err := m.GetSettings(u.ID)
	if err != nil {
		return u, err
	}
	u.Settings = settings
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

func (m *UserModel) GetSettings(userId int64) (UserSettings, error) {
	stmt := `SELECT u.SettingId, a.ItemValue, u.UnconstrainedValue FROM user_settings AS u
			LEFT JOIN allowed_setting_values AS a ON u.SettingId = a.SettingId
			WHERE UserId = ?`
	var settings UserSettings
	rows, err := m.DB.Query(stmt, userId)
	if err != nil {
		return settings, err
	}
	for rows.Next() {
		var id int
		var itemValue sql.NullString
		var unconstrainedValue sql.NullString
		err = rows.Scan(&id, &itemValue, &unconstrainedValue)
		if err != nil {
			return settings, err
		}
		switch id {
		case u_timezone:
			settings.LocalTimezone, err = time.LoadLocation(unconstrainedValue.String)
			if err != nil {
				panic(err)
			}
		}
	}
	return settings, err
}

func (m *UserModel) initializeUserSettings(userId int64, settings UserSettings) error {
	stmt := `INSERT INTO user_settings (UserId, SettingId, AllowedSettingValueId, UnconstrainedValue) VALUES (?, ?, ?, ?)`
	_, err := m.DB.Exec(stmt, userId, u_timezone, nil, settings.LocalTimezone.String())
	if err != nil {
		return err
	}
	return nil
}

func (m *UserModel) UpdateUserSettings(userId int64, settings UserSettings) error {
	err := m.UpdateSetting(userId, m.Settings["local_timezone"], settings.LocalTimezone.String())
	if err != nil {
		return err
	}
	return nil
}

func (m *UserModel) UpdateSetting(userId int64, setting Setting, value string) error {
	stmt := `UPDATE user_settings SET
				AllowedSettingValueId=(SELECT Id FROM allowed_setting_values WHERE SettingId = user_settings.SettingId AND ItemValue = ?),
				UnconstrainedValue=(SELECT ? FROM settings WHERE settings.Id = user_settings.SettingId AND settings.Constrained=0)
			WHERE userId = ?
			AND SettingId = (SELECT Id from Settings WHERE Description=?);`
	result, err := m.DB.Exec(stmt, value, value, userId)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return err
	}
	return nil
}

func (m *UserModel) SetLocalTimezone(userId int64, timezone string) error {
	valid, err := validateSetting(m.DB, u_timezone, timezone)
	if err != nil {
		return err
	}
	if !valid {
		return ErrInvalidSettingValue
	}
	stmt := `UPDATE user_settings SET UnconstrainedValue = ? WHERE UserId = ?`
	_, err = m.DB.Exec(stmt, timezone, userId)
	if err != nil {
		return err
	}
	return nil
}
