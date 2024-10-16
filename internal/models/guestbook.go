package models

import (
	"database/sql"

	"github.com/google/uuid"
)

type Guestbook struct {
    ID uuid.UUID
    SiteUrl string
    UserId uuid.UUID
    IsDeleted bool
    IsActive bool
}

type GuestbookModel struct {
    DB *sql.DB
}

func (m *GuestbookModel) Insert(siteUrl string, userId uuid.UUID) (uuid.UUID, error) {
    id := uuid.New()
    stmt := `INSERT INTO guestbooks (Id, SiteUrl, UserId, IsDeleted, IsActive)
    VALUES(?, ?, FALSE, TRUE)`
    _, err := m.DB.Exec(stmt, id, siteUrl, userId)
    if err != nil {
        return uuid.UUID{}, err
    }
    return id, nil
}

func (m *GuestbookModel) Get(id uuid.UUID) (Guestbook, error) {
    stmt := `SELECT Id, SiteUrl, UserId, IsDeleted, IsActive FROM guestbooks
    WHERE id = ?`
    row := m.DB.QueryRow(stmt, id)
    var g Guestbook
    err := row.Scan(&g.ID, &g.SiteUrl, &g.UserId, &g.IsDeleted, &g.IsActive)
    if err != nil {
        return Guestbook{}, err
    }
    
    return g, nil
}

func (m *GuestbookModel) GetAll(userId uuid.UUID) ([]Guestbook, error) {
    return []Guestbook{}, nil
}
