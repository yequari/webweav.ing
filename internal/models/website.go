package models

import (
	"database/sql"
	"errors"
	"net/url"
	"strconv"
	"time"
)

type Website struct {
	ID      int64
	ShortId uint64
	Name    string
	// SiteUrl    string
	Url        *url.URL
	AuthorName string
	UserId     int64
	Created    time.Time
	Deleted    time.Time
	Guestbook  Guestbook
}

type GuestbookSettings struct {
	IsCommentingEnabled   bool
	ReenableCommenting    time.Time
	IsVisible             bool
	FilteredWords         []string
	AllowRemoteHostAccess bool
}

var ValidDisableDurations = []string{"true", "false", "1h", "4h", "8h", "24h", "72h", "168h"}

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

func (g Guestbook) CanComment() bool {
	now := time.Now().UTC()
	return g.Settings.IsCommentingEnabled && g.Settings.ReenableCommenting.Before(now)
}

type WebsiteModel struct {
	DB       *sql.DB
	Settings map[string]Setting
}

func (m *WebsiteModel) InitializeSettingsMap() error {
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

type WebsiteModelInterface interface {
	Insert(shortId uint64, userId int64, siteName, siteUrl, authorName string) (int64, error)
	Get(shortId uint64) (Website, error)
	GetAllUser(userId int64) ([]Website, error)
	GetAll() ([]Website, error)
	Update(w Website) error
	InitializeSettingsMap() error
	UpdateGuestbookSettings(guestbookId int64, settings GuestbookSettings) error
	UpdateSetting(guestbookId int64, setting Setting, value string) error
	Delete(websiteId int64) error
}

func (m *WebsiteModel) Insert(shortId uint64, userId int64, siteName, siteUrl, authorName string) (int64, error) {
	stmt := `INSERT INTO websites (ShortId, Name, SiteUrl, AuthorName, UserId, Created)
			VALUES (?, ?, ?, ?, ?, ?)`
	tx, err := m.DB.Begin()
	if err != nil {
		return -1, err
	}

	result, err := tx.Exec(stmt, shortId, siteName, siteUrl, authorName, userId, time.Now().UTC())
	// result, err := m.DB.Exec(stmt, shortId, siteName, siteUrl, authorName, userId, time.Now().UTC())
	if err != nil {
		if rollbackError := tx.Rollback(); rollbackError != nil {
			return -1, err
		}
		return -1, err
	}
	websiteId, err := result.LastInsertId()
	if err != nil {
		if rollbackError := tx.Rollback(); rollbackError != nil {
			return -1, err
		}
		return -1, err
	}

	stmt = `INSERT INTO guestbooks (ShortId, UserId, WebsiteId, Created, IsActive)
    VALUES(?, ?, ?, ?, TRUE)`
	result, err = tx.Exec(stmt, shortId, userId, websiteId, time.Now().UTC())
	if err != nil {
		if rollbackError := tx.Rollback(); rollbackError != nil {
			return -1, err
		}
		return -1, err
	}
	guestbookId, err := result.LastInsertId()
	if err != nil {
		if rollbackError := tx.Rollback(); rollbackError != nil {
			return -1, err
		}
		return -1, err
	}

	settings := GuestbookSettings{
		IsCommentingEnabled:   true,
		IsVisible:             true,
		AllowRemoteHostAccess: true,
	}
	stmt = `INSERT INTO guestbook_settings (GuestbookId, SettingId, AllowedSettingValueId, UnconstrainedValue) VALUES 
	(?, ?, ?, ?),
	(?, ?, ?, ?),
	(?, ?, ?, ?),
	(?, ?, ?, ?)`
	_, err = tx.Exec(stmt,
		guestbookId, m.Settings[SettingGbCommentingEnabled].id, settings.IsCommentingEnabled, nil,
		guestbookId, m.Settings[SettingGbReenableComments].id, nil, settings.ReenableCommenting.Format(time.RFC3339),
		guestbookId, m.Settings[SettingGbVisible].id, settings.IsVisible, nil,
		guestbookId, m.Settings[SettingGbAllowRemote].id, settings.AllowRemoteHostAccess, nil)
	if err != nil {
		if rollbackError := tx.Rollback(); rollbackError != nil {
			return -1, err
		}
		return -1, err
	}

	if err := tx.Commit(); err != nil {
		return -1, err
	}
	return websiteId, nil
}

func (m *WebsiteModel) Get(shortId uint64) (Website, error) {
	stmt := `SELECT w.Id, w.ShortId, w.Name, w.SiteUrl, w.AuthorName, w.UserId, w.Created
	FROM websites AS w
	WHERE w.ShortId = ? AND w.DELETED IS NULL`
	tx, err := m.DB.Begin()
	if err != nil {
		return Website{}, nil
	}
	row := tx.QueryRow(stmt, shortId)
	var w Website
	var u string
	err = row.Scan(&w.ID, &w.ShortId, &w.Name, &u, &w.AuthorName, &w.UserId, &w.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrNoRecord
		}
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return Website{}, err
		}
		return Website{}, err
	}
	w.Url, err = url.Parse(u)
	if err != nil {
		return Website{}, err
	}

	stmt = `SELECT Id, ShortId, UserId, WebsiteId, Created, IsActive FROM guestbooks
    WHERE WebsiteId = ? AND Deleted IS NULL`
	row = tx.QueryRow(stmt, w.ID)
	var g Guestbook
	err = row.Scan(&g.ID, &g.ShortId, &g.UserId, &g.WebsiteId, &g.Created, &g.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrNoRecord
		}
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return Website{}, err
		}
		return Website{}, err
	}

	gbSettings, err := m.getGuestbookSettings(tx, g.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrNoRecord
		}
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return Website{}, err
		}
		return Website{}, err
	}

	err = tx.Commit()
	if err != nil {
		return Website{}, nil
	}

	// if comment disable setting has expired, enable commenting
	commentingReenabled := time.Now().UTC().After(gbSettings.ReenableCommenting)
	if commentingReenabled {
		gbSettings.IsCommentingEnabled = true
	}

	g.Settings = gbSettings
	w.Guestbook = g
	return w, nil
}

func (m *WebsiteModel) GetAllUser(userId int64) ([]Website, error) {
	stmt := `SELECT w.Id, w.ShortId, w.Name, w.SiteUrl, w.AuthorName, w.UserId, w.Created, 
	g.Id, g.ShortId, g.Created, g.IsActive
	FROM websites AS w INNER JOIN guestbooks AS g ON w.Id = g.WebsiteId
	WHERE w.UserId = ? AND w.Deleted IS NULL`
	rows, err := m.DB.Query(stmt, userId)
	if err != nil {
		return nil, err
	}
	var websites []Website
	for rows.Next() {
		var w Website
		var u string
		err := rows.Scan(&w.ID, &w.ShortId, &w.Name, &u, &w.AuthorName, &w.UserId, &w.Created,
			&w.Guestbook.ID, &w.Guestbook.ShortId, &w.Guestbook.Created, &w.Guestbook.IsActive)
		if err != nil {
			return nil, err
		}
		w.Url, err = url.Parse(u)
		if err != nil {
			return nil, err
		}
		websites = append(websites, w)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return websites, nil
}

func (m *WebsiteModel) GetAll() ([]Website, error) {
	stmt := `SELECT w.Id, w.ShortId, w.Name, w.SiteUrl, w.AuthorName, w.UserId, w.Created, 
	g.Id, g.ShortId, g.Created, g.IsActive
	FROM websites AS w INNER JOIN guestbooks AS g ON w.Id = g.WebsiteId WHERE w.Deleted IS NULL`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	var websites []Website
	for rows.Next() {
		var w Website
		var u string
		err := rows.Scan(&w.ID, &w.ShortId, &w.Name, &u, &w.AuthorName, &w.UserId, &w.Created,
			&w.Guestbook.ID, &w.Guestbook.ShortId, &w.Guestbook.Created, &w.Guestbook.IsActive)
		if err != nil {
			return nil, err
		}
		w.Url, err = url.Parse(u)
		if err != nil {
			return nil, err
		}
		websites = append(websites, w)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return websites, nil
}

func (m *WebsiteModel) Update(w Website) error {
	stmt := `UPDATE websites SET Name = ?, SiteUrl = ?, AuthorName = ? WHERE ID = ?`
	r, err := m.DB.Exec(stmt, w.Name, w.Url.String(), w.AuthorName, w.ID)
	if err != nil {
		return err
	}
	if rows, err := r.RowsAffected(); rows != 1 {
		if err != nil {
			return err
		}
		return errors.New("Failed to update website")
	}
	return nil
}

func (m *WebsiteModel) Delete(websiteId int64) error {
	stmt := `UPDATE websites SET Deleted = ? WHERE ID = ?`
	r, err := m.DB.Exec(stmt, time.Now().UTC(), websiteId)
	if err != nil {
		return err
	}
	if rows, err := r.RowsAffected(); rows != 1 {
		if err != nil {
			return err
		}
		return errors.New("Failed to update website")
	}
	return nil
}

func (m *WebsiteModel) getGuestbookSettings(tx *sql.Tx, guestbookId int64) (GuestbookSettings, error) {
	stmt := `SELECT g.SettingId, a.ItemValue, g.UnconstrainedValue FROM guestbook_settings AS g
			LEFT JOIN allowed_setting_values AS a ON g.AllowedSettingValueId = a.Id
			WHERE GuestbookId = ?`
	var settings GuestbookSettings
	rows, err := tx.Query(stmt, guestbookId)
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
		case m.Settings[SettingGbCommentingEnabled].id:
			settings.IsCommentingEnabled, err = strconv.ParseBool(itemValue.String)
			if err != nil {
				return settings, err
			}
			break
		case m.Settings[SettingGbReenableComments].id:
			settings.ReenableCommenting, err = time.Parse(time.RFC3339, unconstrainedValue.String)
			if err != nil {
				return settings, err
			}
			break
		case m.Settings[SettingGbVisible].id:
			settings.IsVisible, err = strconv.ParseBool(itemValue.String)
			if err != nil {
				return settings, err
			}
			break
		case m.Settings[SettingGbAllowRemote].id:
			settings.AllowRemoteHostAccess, err = strconv.ParseBool(itemValue.String)
			if err != nil {
				return settings, err
			}
			break
		}
	}
	return settings, nil
}

func (m *WebsiteModel) UpdateGuestbookSettings(guestbookId int64, settings GuestbookSettings) error {
	err := m.UpdateSetting(guestbookId, m.Settings[SettingGbVisible], strconv.FormatBool(settings.IsVisible))
	if err != nil {
		return err
	}
	err = m.UpdateSetting(guestbookId, m.Settings[SettingGbAllowRemote], strconv.FormatBool(settings.AllowRemoteHostAccess))
	if err != nil {
		return err
	}
	err = m.UpdateSetting(guestbookId, m.Settings[SettingGbCommentingEnabled], strconv.FormatBool(settings.IsCommentingEnabled))
	if err != nil {
		return err
	}
	err = m.UpdateSetting(guestbookId, m.Settings[SettingGbReenableComments], settings.ReenableCommenting.Format(time.RFC3339))
	if err != nil {
		return err
	}
	return nil
}

func (m *WebsiteModel) UpdateSetting(guestbookId int64, setting Setting, value string) error {
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
