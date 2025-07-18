package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type UserSettings struct {
	LocalTimezone *time.Location
}

const (
	SettingUserTimezone = "local_timezone"
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
	Settings       UserSettings
}

type UserModel struct {
	DB       *sql.DB
	Settings map[string]Setting
}

type UserIdToken struct {
	Subject       string   `json:"sub"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Username      string   `json:"preferred_username"`
	Groups        []string `json:"groups"`
}

type UserModelInterface interface {
	InitializeSettingsMap() error
	Insert(shortId uint64, username string, email string, password string, settings UserSettings) error
	InsertWithoutPassword(shortId uint64, username string, email string, subject string, settings UserSettings) (int64, error)
	Get(shortId uint64) (User, error)
	GetById(id int64) (User, error)
	GetByEmail(email string) (int64, error)
	GetBySubject(subject string) (int64, error)
	GetAll() ([]User, error)
	Authenticate(email, password string) (int64, error)
	Exists(id int64) (bool, error)
	UpdateUserSettings(userId int64, settings UserSettings) error
	UpdateSetting(userId int64, setting Setting, value string) error
	UpdateSubject(userId int64, subject string) error
}

func (m *UserModel) InitializeSettingsMap() error {
	if m.Settings == nil {
		m.Settings = make(map[string]Setting)
	}
	stmt := `SELECT settings.Id, settings.Description, Constrained, d.Id, d.Description, g.Id, g.Description, MinValue, MaxValue
        FROM settings
        LEFT JOIN setting_data_types d ON settings.DataType = d.Id
        LEFT JOIN setting_groups g ON settings.SettingGroup = g.Id
        WHERE SettingGroup = (SELECT Id FROM setting_groups WHERE Description = 'user' LIMIT 1)`
	result, err := m.DB.Query(stmt)
	if err != nil {
		return err
	}
	for result.Next() {
		var s Setting
		var mn sql.NullString
		var mx sql.NullString
		err := result.Scan(&s.id, &s.description, &s.constrained, &s.dataType.id, &s.dataType.description, &s.settingGroup.id, &s.settingGroup.description, &mn, &mx)
		if mn.Valid {
			s.minValue = mn.String
		}
		if mx.Valid {
			s.maxValue = mx.String
		}
		if err != nil {
			return err
		}
		m.Settings[s.description] = s
	}
	return nil
}

func (m *UserModel) Insert(shortId uint64, username string, email string, password string, settings UserSettings) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	stmt := `INSERT INTO users (ShortId, Username, Email, IsBanned, HashedPassword, Created)
    VALUES (?, ?, ?, FALSE, ?, ?)`
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	result, err := tx.Exec(stmt, shortId, username, email, hashedPassword, time.Now().UTC())
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return err
		}
		if sqliteError, ok := err.(sqlite3.Error); ok {
			if sqliteError.ExtendedCode == 2067 && strings.Contains(sqliteError.Error(), "Email") {
				return ErrDuplicateEmail
			}
		}
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return err
		}
		return err
	}
	stmt = `INSERT INTO user_settings (UserId, SettingId, AllowedSettingValueId, UnconstrainedValue) 
		VALUES (?, ?, ?, ?)`
	_, err = tx.Exec(stmt, id, m.Settings[SettingUserTimezone].id, nil, settings.LocalTimezone.String())
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return err
		}
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (m *UserModel) InsertWithoutPassword(shortId uint64, username string, email string, subject string, settings UserSettings) (int64, error) {
	stmt := `INSERT INTO users (ShortId, Username, Email, IsBanned, OIDCSubject, Created)
    VALUES (?, ?, ?, FALSE, ?, ?)`
	tx, err := m.DB.Begin()
	if err != nil {
		return -1, err
	}
	result, err := tx.Exec(stmt, shortId, username, email, subject, time.Now().UTC())
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return -1, err
		}
		if sqliteError, ok := err.(sqlite3.Error); ok {
			if sqliteError.ExtendedCode == 2067 && strings.Contains(sqliteError.Error(), "Email") {
				return -1, ErrDuplicateEmail
			}
		}
		return -1, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return -1, err
		}
		return -1, err
	}
	stmt = `INSERT INTO user_settings (UserId, SettingId, AllowedSettingValueId, UnconstrainedValue) 
		VALUES (?, ?, ?, ?)`
	_, err = tx.Exec(stmt, id, m.Settings[SettingUserTimezone].id, nil, settings.LocalTimezone.String())
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return -1, err
		}
		return -1, err
	}
	err = tx.Commit()
	if err != nil {
		return -1, err
	}

	return id, nil
}

func (m *UserModel) Get(shortId uint64) (User, error) {
	stmt := `SELECT Id, ShortId, Username, Email, Created FROM users WHERE ShortId = ? AND Deleted IS NULL`
	tx, err := m.DB.Begin()
	if err != nil {
		return User{}, err
	}
	row := tx.QueryRow(stmt, shortId)
	var u User
	err = row.Scan(&u.ID, &u.ShortId, &u.Username, &u.Email, &u.Created)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return User{}, err
		}
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNoRecord
		}
		return User{}, err
	}
	settings, err := m.getSettings(tx, u.ID)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return User{}, err
		}
		return User{}, err
	}
	u.Settings = settings
	err = tx.Commit()
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (m *UserModel) GetById(id int64) (User, error) {
	stmt := `SELECT Id, ShortId, Username, Email, Created FROM users WHERE Id = ? AND Deleted IS NULL`
	tx, err := m.DB.Begin()
	if err != nil {
		return User{}, err
	}
	row := tx.QueryRow(stmt, id)
	var u User
	err = row.Scan(&u.ID, &u.ShortId, &u.Username, &u.Email, &u.Created)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return User{}, err
		}
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNoRecord
		}
		return User{}, err
	}
	settings, err := m.getSettings(tx, u.ID)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return User{}, err
		}
		return User{}, err
	}
	u.Settings = settings
	err = tx.Commit()
	if err != nil {
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

func (m *UserModel) GetByEmail(email string) (int64, error) {
	var id int64
	stmt := `SELECT Id FROM users WHERE Email = ?`
	err := m.DB.QueryRow(stmt, email).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return -1, ErrNoRecord
		} else {
			return -1, err
		}
	}
	return id, nil
}

func (m *UserModel) GetBySubject(subject string) (int64, error) {
	var id int64
	var s sql.NullString
	stmt := `SELECT Id, OIDCSubject FROM users WHERE OIDCSubject = ?`
	err := m.DB.QueryRow(stmt, subject).Scan(&id, &s)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return -1, ErrNoRecord
		} else {
			return -1, err
		}
	}
	return id, nil
}

func (m *UserModel) UpdateSubject(userId int64, subject string) error {
	stmt := `UPDATE users SET OIDCSubject = ? WHERE Id = ?`
	_, err := m.DB.Exec(stmt, subject, userId)
	if err != nil {
		return err
	}
	return nil
}

func (m *UserModel) Exists(id int64) (bool, error) {
	var exists bool
	stmt := `SELECT EXISTS(SELECT true FROM users WHERE Id = ? AND DELETED IS NULL)`
	err := m.DB.QueryRow(stmt, id).Scan(&exists)
	return exists, err
}

func (m *UserModel) getSettings(tx *sql.Tx, userId int64) (UserSettings, error) {
	stmt := `SELECT u.SettingId, a.ItemValue, u.UnconstrainedValue FROM user_settings AS u
			LEFT JOIN allowed_setting_values AS a ON u.AllowedSettingValueId = a.Id
			WHERE UserId = ?`
	var settings UserSettings
	rows, err := tx.Query(stmt, userId)
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
		case m.Settings[SettingUserTimezone].id:
			settings.LocalTimezone, err = time.LoadLocation(unconstrainedValue.String)
			if err != nil {
				panic(err)
			}
		}
	}
	return settings, err
}

func (m *UserModel) UpdateUserSettings(userId int64, settings UserSettings) error {
	err := m.UpdateSetting(userId, m.Settings[SettingUserTimezone], settings.LocalTimezone.String())
	if err != nil {
		return err
	}
	return nil
}

func (m *UserModel) UpdateSetting(userId int64, setting Setting, value string) error {
	valid := setting.Validate(value)
	if !valid {
		return ErrInvalidSettingValue
	}
	stmt := `UPDATE user_settings SET
				AllowedSettingValueId=IFNULL(
					(SELECT Id FROM allowed_setting_values WHERE SettingId = user_settings.SettingId AND ItemValue = ?), AllowedSettingValueId
				),
				UnconstrainedValue=(SELECT ? FROM settings WHERE settings.Id = user_settings.SettingId AND settings.Constrained=0)
			WHERE userId = ?
			AND SettingId = (SELECT Id from Settings WHERE Description=?);`
	result, err := m.DB.Exec(stmt, value, value, userId, setting.description)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrInvalidSettingValue
	}
	return nil
}
