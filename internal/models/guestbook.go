package models

import (
	"database/sql"
	"strconv"
	"time"
)

type Guestbook struct {
    ID int64
    ShortId uint64
    SiteUrl string
    UserId int64
    Created time.Time
    IsDeleted bool
    IsActive bool
}

func (gb Guestbook) Slug() string {
    return strconv.FormatUint(gb.ShortId, 36)
}

type GuestbookModel struct {
    DB *sql.DB
}

func (m *GuestbookModel) Insert(shortId uint64, siteUrl string, userId int64) (int64, error) {
    stmt := `INSERT INTO guestbooks (ShortId, SiteUrl, UserId, Created, IsDeleted, IsActive)
    VALUES(?, ?, ?, ?, FALSE, TRUE)`
    result, err := m.DB.Exec(stmt, shortId, siteUrl, userId, time.Now().UTC())
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
    stmt := `SELECT Id, ShortId, SiteUrl, UserId, Created, IsDeleted, IsActive FROM guestbooks
    WHERE ShortId = ?`
    row := m.DB.QueryRow(stmt, shortId)
    var g Guestbook
    err := row.Scan(&g.ID, &g.ShortId, &g.SiteUrl, &g.UserId, &g.Created, &g.IsDeleted, &g.IsActive)
    if err != nil {
        return Guestbook{}, err
    }
    
    return g, nil
}

func (m *GuestbookModel) GetAll(userId int64) ([]Guestbook, error) {
    stmt := `SELECT Id, ShortId, SiteUrl, UserId, Created, IsDeleted, IsActive FROM guestbooks
    WHERE UserId = ?`
    rows, err := m.DB.Query(stmt, userId)
    if err != nil {
        return nil, err
    }
    var guestbooks []Guestbook
    for rows.Next() {
        var g Guestbook
        err = rows.Scan(&g.ID, &g.ShortId, &g.SiteUrl, &g.UserId, &g.Created, &g.IsDeleted, &g.IsActive)
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
