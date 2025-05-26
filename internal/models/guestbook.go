package models

import (
	"database/sql"
	"strconv"
	"time"
)

type GuestbookSettings struct {
	IsCommentingEnabled   bool
	ReenableCommenting    time.Time
	IsVisible             bool
	FilteredWords         []string
	AllowRemoteHostAccess bool
}

const (
	SettingGbCommentingEnabled = "commenting_enabled"
	SettingGbReenableComments  = "reenable_comments"
	SettingGbVisible           = "is_visible"
	SettingGbFilteredWords     = "filtered_words"
	SettingGbAllowRemote       = "remote_enabled"
)

type Guestbook struct {
	ID        int64
	ShortId   uint64
	UserId    int64
	WebsiteId int64
	Created   time.Time
	Deleted   time.Time
	IsActive  bool
	Settings  GuestbookSettings
}

type GuestbookModel struct {
	DB       *sql.DB
	Settings map[string]Setting
}

func (m *GuestbookModel) InitializeSettingsMap() error {
	if m.Settings == nil {
		m.Settings = make(map[string]Setting)
	}
	stmt := `SELECT settings.Id, settings.Description, Constrained, d.Id, d.Description, g.Id, g.Description, MinValue, MaxValue
        FROM settings
        LEFT JOIN setting_data_types d ON settings.DataType = d.Id
        LEFT JOIN setting_groups g ON settings.SettingGroup = g.Id
        WHERE SettingGroup = (SELECT Id FROM setting_groups WHERE Description = 'guestbook' LIMIT 1)`
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

func (m *GuestbookModel) Insert(shortId uint64, userId int64, websiteId int64, settings GuestbookSettings) (int64, error) {
	stmt := `INSERT INTO guestbooks (ShortId, UserId, WebsiteId, Created, IsActive)
    VALUES(?, ?, ?, ?, TRUE)`
	result, err := m.DB.Exec(stmt, shortId, userId, websiteId, time.Now().UTC())
	if err != nil {
		return -1, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}
	err = m.initializeGuestbookSettings(id, settings)
	if err != nil {
		return id, err
	}
	return id, nil
}

func (m *GuestbookModel) Get(shortId uint64) (Guestbook, error) {
	stmt := `SELECT Id, ShortId, UserId, WebsiteId, Created, Deleted, IsActive FROM guestbooks
    WHERE ShortId = ?`
	row := m.DB.QueryRow(stmt, shortId)
	var g Guestbook
	var t sql.NullTime
	err := row.Scan(&g.ID, &g.ShortId, &g.UserId, &g.WebsiteId, &g.Created, &t, &g.IsActive)
	if err != nil {
		return Guestbook{}, err
	}
	if t.Valid {
		g.Deleted = t.Time
	}
	return g, nil
}

func (m *GuestbookModel) GetAll(userId int64) ([]Guestbook, error) {
	stmt := `SELECT Id, ShortId, UserId, WebsiteId, Created, IsActive FROM guestbooks
    WHERE UserId = ? AND DELETED IS NULL`
	rows, err := m.DB.Query(stmt, userId)
	if err != nil {
		return nil, err
	}
	var guestbooks []Guestbook
	for rows.Next() {
		var g Guestbook
		err = rows.Scan(&g.ID, &g.ShortId, &g.UserId, &g.WebsiteId, &g.Created, &g.IsActive)
		if err != nil {
			return nil, err
		}
		guestbooks = append(guestbooks, g)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return guestbooks, nil
}

func (m *GuestbookModel) initializeGuestbookSettings(guestbookId int64, settings GuestbookSettings) error {
	stmt := `INSERT INTO guestbook_settings (GuestbookId, SettingId, AllowedSettingValueId, UnconstrainedValue) VALUES 
	(?, ?, ?, ?),
	(?, ?, ?, ?),
	(?, ?, ?, ?),
	(?, ?, ?, ?),
	(?, ?, ?, ?)`
	_, err := m.DB.Exec(stmt,
		guestbookId, m.Settings[SettingGbCommentingEnabled].id, settings.IsCommentingEnabled, nil,
		guestbookId, m.Settings[SettingGbReenableComments].id, nil, settings.ReenableCommenting.String(),
		guestbookId, m.Settings[SettingGbVisible].id, settings.IsVisible, nil,
		guestbookId, m.Settings[SettingGbAllowRemote].id, settings.AllowRemoteHostAccess, nil)
	if err != nil {
		return err
	}
	return nil
}

func (m *GuestbookModel) UpdateGuestbookSettings(guestbookId int64, settings GuestbookSettings) error {
	err := m.UpdateSetting(guestbookId, m.Settings[SettingGbCommentingEnabled], strconv.FormatBool(settings.IsCommentingEnabled))
	if err != nil {
		return err
	}
	err = m.UpdateSetting(guestbookId, m.Settings[SettingGbReenableComments], settings.ReenableCommenting.String())
	if err != nil {
		return err
	}
	return nil
}

func (m *GuestbookModel) UpdateSetting(guestbookId int64, setting Setting, value string) error {
	stmt := `UPDATE guestbook_settings SET
				AllowedSettingValueId=IFNULL(
					(SELECT Id FROM allowed_setting_values WHERE SettingId = guestbook_settings.SettingId AND ItemValue = ?), AllowedSettingValueId
				), 
				UnconstrainedValue=(SELECT ? FROM settings WHERE settings.Id = guestbook_settings.SettingId AND settings.Constrained=0)
			WHERE GuestbookId = ?
			AND SettingId = (SELECT Id from Settings WHERE Description=?);`
	result, err := m.DB.Exec(stmt, value, value, guestbookId, setting.description)
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

func (m *GuestbookModel) AddFilteredWord(guestbookId int64, word string) error {
	return nil
}

func (m *GuestbookModel) RemoveFilteredWord(guestbookId int64, word string) error {
	return nil
}
