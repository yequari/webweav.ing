package models

import (
	"database/sql"
	"time"
)

type GuestbookComment struct {
	ID          int64
	ShortId     uint64
	GuestbookId int64
	ParentId    int64
	AuthorName  string
	AuthorEmail string
	AuthorSite  string
	CommentText string
	PageUrl     string
	Created     time.Time
	Deleted     time.Time
	IsPublished bool
}

type GuestbookCommentModel struct {
	DB *sql.DB
}

func (m *GuestbookCommentModel) Insert(shortId uint64, guestbookId, parentId int64, authorName,
	authorEmail, authorSite, commentText, pageUrl string, isPublished bool) (int64, error) {
	stmt := `INSERT INTO guestbook_comments (ShortId, GuestbookId, ParentId, AuthorName,
    AuthorEmail, AuthorSite, CommentText, PageUrl, Created, IsPublished)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := m.DB.Exec(stmt, shortId, guestbookId, parentId, authorName, authorEmail,
		authorSite, commentText, pageUrl, time.Now().UTC(), isPublished)
	if err != nil {
		return -1, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (m *GuestbookCommentModel) Get(shortId uint64) (GuestbookComment, error) {
	stmt := `SELECT Id, ShortId, GuestbookId, ParentId, AuthorName, AuthorEmail, AuthorSite,
    CommentText, PageUrl, Created, IsPublished, Deleted FROM guestbook_comments WHERE ShortId = ?`
	row := m.DB.QueryRow(stmt, shortId)
	var c GuestbookComment
	var t sql.NullTime
	err := row.Scan(&c.ID, &c.ShortId, &c.GuestbookId, &c.ParentId, &c.AuthorName, &c.AuthorEmail, &c.AuthorSite,
		&c.CommentText, &c.PageUrl, &c.Created, &c.IsPublished, &t)
	if err != nil {
		return GuestbookComment{}, err
	}
	if t.Valid {
		c.Deleted = t.Time
	}
	return c, nil
}

func (m *GuestbookCommentModel) GetAll(guestbookId int64) ([]GuestbookComment, error) {
	stmt := `SELECT Id, ShortId, GuestbookId, ParentId, AuthorName, AuthorEmail, AuthorSite,
    CommentText, PageUrl, Created, IsPublished 
	    FROM guestbook_comments 
	    WHERE GuestbookId = ? AND IsPublished = TRUE AND DELETED IS NULL
	    ORDER BY Created DESC`
	rows, err := m.DB.Query(stmt, guestbookId)
	if err != nil {
		return nil, err
	}
	var comments []GuestbookComment
	for rows.Next() {
		var c GuestbookComment
		err = rows.Scan(&c.ID, &c.ShortId, &c.GuestbookId, &c.ParentId, &c.AuthorName, &c.AuthorEmail, &c.AuthorSite,
			&c.CommentText, &c.PageUrl, &c.Created, &c.IsPublished)
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

func (m *GuestbookCommentModel) GetDeleted(guestbookId int64) ([]GuestbookComment, error) {
	stmt := `SELECT Id, ShortId, GuestbookId, ParentId, AuthorName, AuthorEmail, AuthorSite,
    CommentText, PageUrl, Created, IsPublished, Deleted
	    FROM guestbook_comments 
	    WHERE GuestbookId = ? AND Deleted IS NOT NULL
	    ORDER BY Created DESC`
	rows, err := m.DB.Query(stmt, guestbookId)
	if err != nil {
		return nil, err
	}
	var comments []GuestbookComment
	for rows.Next() {
		var c GuestbookComment
		var t sql.NullTime
		err = rows.Scan(&c.ID, &c.ShortId, &c.GuestbookId, &c.ParentId, &c.AuthorName, &c.AuthorEmail, &c.AuthorSite,
			&c.CommentText, &c.PageUrl, &c.Created, &c.IsPublished, &t)
		if err != nil {
			return nil, err
		}
		if t.Valid {
			c.Deleted = t.Time
		}
		comments = append(comments, c)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return comments, nil
}

func (m *GuestbookCommentModel) GetUnpublished(guestbookId int64) ([]GuestbookComment, error) {
	stmt := `SELECT Id, ShortId, GuestbookId, ParentId, AuthorName, AuthorEmail, AuthorSite,
    CommentText, PageUrl, Created, IsPublished 
	    FROM guestbook_comments 
	    WHERE GuestbookId = ? AND IsDeleted IS NULL AND IsPublished = FALSE
	    ORDER BY Created DESC`
	rows, err := m.DB.Query(stmt, guestbookId)
	if err != nil {
		return nil, err
	}
	var comments []GuestbookComment
	for rows.Next() {
		var c GuestbookComment
		err = rows.Scan(&c.ID, &c.ShortId, &c.GuestbookId, &c.ParentId, &c.AuthorName, &c.AuthorEmail, &c.AuthorSite,
			&c.CommentText, &c.PageUrl, &c.Created, &c.IsPublished)
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

func (m *GuestbookCommentModel) UpdateComment(comment *GuestbookComment) error {
	stmt := `UPDATE guestbook_comments
			SET CommentText = ?,
				PageUrl = ?,
				IsPublished = ?,
				Deleted = ?
		WHERE Id = ?`
	_, err := m.DB.Exec(stmt, comment.CommentText, comment.PageUrl, comment.IsPublished, comment.Deleted, comment.ID)
	if err != nil {
		return err
	}
	return nil
}
