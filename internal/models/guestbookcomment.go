package models

import (
	"database/sql"

	"github.com/google/uuid"
)

type GuestbookComment struct {
    ID uuid.UUID
    GuestbookId uuid.UUID
    ParentId uuid.UUID
    AuthorName string
    AuthorEmail string
    AuthorSite string
    CommentText string
    PageUrl string
    IsPublished bool
    IsDeleted bool
}

type GuestbookCommentModel struct {
    DB *sql.DB
}

func (m *GuestbookCommentModel) Insert(guestbookId, parentId uuid.UUID, authorName,
    authorEmail, authorSite, commentText, pageUrl string, isPublished bool) (uuid.UUID, error) {
    id := uuid.New()
    stmt := `INSERT INTO guestbook_comments (Id, GuestbookId, ParentId, AuthorName,
    AuthorEmail, AuthorSite, CommentText, PageUrl, IsPublished, IsDeleted)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, FALSE)`
    _, err := m.DB.Exec(stmt, id, guestbookId, parentId, authorName, authorEmail,
	authorSite, commentText, pageUrl, isPublished)
    if err != nil {
	return uuid.UUID{}, err
    }
    return id, nil
}

func (m *GuestbookCommentModel) Get(id uuid.UUID) (GuestbookComment, error) {
    stmt := `SELECT Id, GuestbookId, ParentId, AuthorName, AuthorEmail, AuthorSite,
    CommentText, PageUrl, IsPublished, IsDeleted FROM guestbook_comments WHERE id = ?`
    row := m.DB.QueryRow(stmt, id)
    var c GuestbookComment
    err := row.Scan(&c.ID, &c.GuestbookId, &c.ParentId, &c.AuthorName, &c.AuthorEmail, &c.AuthorSite, &c.CommentText, &c.PageUrl, &c.IsPublished, &c.IsDeleted)
    if err != nil {
	return GuestbookComment{}, err
    }
    return c, nil
}

func (m *GuestbookCommentModel) GetAll(guestbookId uuid.UUID) ([]GuestbookComment, error) {
    stmt := `SELECT Id, GuestbookId, ParentId, AuthorName, AuthorEmail, AuthorSite,
    CommentText, PageUrl, IsPublished, IsDeleted FROM guestbook_comments WHERE GuestbookId = ?`
    rows, err := m.DB.Query(stmt, guestbookId)
    if err != nil {
	return nil, err
    }
    var comments []GuestbookComment
    for rows.Next() {
	var c GuestbookComment
	err = rows.Scan(&c.ID, &c.GuestbookId, &c.ParentId, &c.AuthorName, &c.AuthorEmail, &c.AuthorSite, &c.CommentText, &c.PageUrl, &c.IsPublished, &c.IsDeleted)
	if err != nil {
	    return nil, err
	}
	comments = append(comments, c)
    }
    if err = rows.Err(); err != nil {
	return nil, err
    }
    return comments, nil
}
