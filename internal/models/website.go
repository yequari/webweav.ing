package models

import (
	"database/sql"
	"time"
)

type Website struct {
	ID         int64
	ShortId    uint64
	Name       string
	SiteUrl    string
	AuthorName string
	UserId     int64
	Created    time.Time
	Deleted    time.Time
	Guestbook  Guestbook
}

type WebsiteModel struct {
	DB *sql.DB
}

func (m *WebsiteModel) Insert(shortId uint64, userId int64, siteName, siteUrl, authorName string) (int64, error) {
	stmt := `INSERT INTO websites (ShortId, Name, SiteUrl, AuthorName, UserId, Created)
			VALUES (?, ?, ?, ?, ?, ?)`
	result, err := m.DB.Exec(stmt, shortId, siteName, siteUrl, authorName, userId, time.Now().UTC())
	if err != nil {
		return -1, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (m *WebsiteModel) Get(shortId uint64) (Website, error) {
	stmt := `SELECT w.Id, w.ShortId, w.Name, w.SiteUrl, w.AuthorName, w.UserId, w.Created, w.Deleted,
	g.Id, g.ShortId, g.Created, g.IsDeleted, g.IsActive
	FROM websites AS w INNER JOIN guestbooks AS g ON w.Id = g.WebsiteId
	WHERE w.ShortId = ?`
	row := m.DB.QueryRow(stmt, shortId)
	var t sql.NullTime
	var w Website
	err := row.Scan(&w.ID, &w.ShortId, &w.Name, &w.SiteUrl, &w.AuthorName, &w.UserId, &w.Created, &t,
		&w.Guestbook.ID, &w.Guestbook.ShortId, &w.Guestbook.Created, &w.Guestbook.IsDeleted, &w.Guestbook.IsActive)
	if err != nil {
		return Website{}, err
	}
	// handle if Deleted is null
	if t.Valid {
		w.Deleted = t.Time
	}
	return w, nil
}

func (m *WebsiteModel) GetById(id int64) (Website, error) {
	stmt := `SELECT w.Id, w.ShortId, w.Name, w.SiteUrl, w.AuthorName, w.UserId, w.Created, w.Deleted,
	g.Id, g.ShortId, g.Created, g.IsDeleted, g.IsActive
	FROM websites AS w INNER JOIN guestbooks AS g ON w.Id = g.WebsiteId
	WHERE w.Id = ?`
	row := m.DB.QueryRow(stmt, id)
	var t sql.NullTime
	var w Website
	err := row.Scan(&w.ID, &w.ShortId, &w.Name, &w.SiteUrl, &w.AuthorName, &w.UserId, &w.Created, &t,
		&w.Guestbook.ID, &w.Guestbook.ShortId, &w.Guestbook.Created, &w.Guestbook.IsDeleted, &w.Guestbook.IsActive)
	if err != nil {
		return Website{}, err
	}
	// handle if Deleted is null
	if t.Valid {
		w.Deleted = t.Time
	}
	return w, nil
}

func (m *WebsiteModel) GetAll(userId int64) ([]Website, error) {
	stmt := `SELECT w.Id, w.ShortId, w.Name, w.SiteUrl, w.AuthorName, w.UserId, w.Created, w.Deleted,
	g.Id, g.ShortId, g.Created, g.IsDeleted, g.IsActive
	FROM websites AS w INNER JOIN guestbooks AS g ON w.Id = g.WebsiteId
	WHERE w.UserId = ?`
	rows, err := m.DB.Query(stmt, userId)
	if err != nil {
		return nil, err
	}
	var websites []Website
	for rows.Next() {
		var t sql.NullTime
		var w Website
		err := rows.Scan(&w.ID, &w.ShortId, &w.Name, &w.SiteUrl, &w.AuthorName, &w.UserId, &w.Created, &t,
			&w.Guestbook.ID, &w.Guestbook.ShortId, &w.Guestbook.Created, &w.Guestbook.IsDeleted, &w.Guestbook.IsActive)
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
