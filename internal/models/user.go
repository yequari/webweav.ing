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
	if err != nil {
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
			LEFT JOIN allowed_setting_values AS a ON u.AllowedSettingValueId = a.Id
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
		case m.Settings[SettingUserTimezone].id:
			settings.LocalTimezone, err = time.LoadLocation(unconstrainedValue.String)
			if err != nil {
				panic(err)
			}
		}
	}
	return settings, err
}

func (m *UserModel) initializeUserSettings(userId int64, settings UserSettings) error {
	stmt := `INSERT INTO user_settings (UserId, SettingId, AllowedSettingValueId, UnconstrainedValue) 
		VALUES (?, ?, ?, ?)`
	_, err := m.DB.Exec(stmt, userId, m.Settings[SettingUserTimezone].id, nil, settings.LocalTimezone.String())
	if err != nil {
		return err
	}
	return nil
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
