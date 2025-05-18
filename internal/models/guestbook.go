package models

import (
	"database/sql"
	"time"
)

type GuestbookSettings struct {
	IsCommentingEnabled   bool
	ReenableCommenting    time.Time
	IsVisible             bool
	FilteredWords         []string
	AllowRemoteHostAccess bool
}

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
	DB *sql.DB
}

func (m *GuestbookModel) Insert(shortId uint64, userId int64, websiteId int64) (int64, error) {
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

func (m *GuestbookModel) initializeGuestbookSettings(guestbookId int64, settings UserSettings) error {
	stmt := `INSERT INTO guestbook_settings (GuestbookId, SettingId, AllowedSettingValueId, UnconstrainedValue) VALUES 
	(?, ?, ?, ?),
	(?, ?, ?, ?),
	(?, ?, ?, ?),
	(?, ?, ?, ?),
	(?, ?, ?, ?)`
	_ = len(stmt)
	return nil
}

func (m *GuestbookModel) UpdateSetting(guestbookId int64, value string) {
	stmt := `UPDATE guestbook_settings SET
		AllowedSettingValueId=IFNULL((SELECT Id FROM allowed_setting_values WHERE SettingId = guestbook_settings.SettingId AND ItemValue = ?), AllowedSettingValueId), 
		UnconstrainedValue=(SELECT ? FROM settings WHERE settings.Id = guestbook_settings.SettingId AND settings.Constrained=0)
	WHERE GuestbookId = ?
		AND SettingId = (SELECT Id from Settings WHERE Description=?);`
	_ = len(stmt)
}

func (m *GuestbookModel) SetCommentingEnabled(guestbookId int64, enabled bool) error {
	return nil
}

func (m *GuestbookModel) SetReenableCommentingDate(guestbookId int64, reenableTime time.Time) error {
	return nil
}

func (m *GuestbookModel) SetVisible(guestbookId int64, visible bool) error {
	return nil
}

func (m *GuestbookModel) AddFilteredWord(guestbookId int64, word string) error {
	return nil
}

func (m *GuestbookModel) RemoveFilteredWord(guestbookId int64, word string) error {
	return nil
}

func (m *GuestbookModel) SetRemoteHostAccess(guestbookId int64, allowed bool) error {
	return nil
}
