package mocks

import (
	"time"

	"git.32bit.cafe/32bitcafe/guestbook/internal/models"
)

var mockGuestbookComment = models.GuestbookComment{
	ID:          1,
	ShortId:     1,
	GuestbookId: 1,
	AuthorName:  "John Test",
	AuthorEmail: "test@example.com",
	AuthorSite:  "example.com",
	CommentText: "Hello, world",
	Created:     time.Now(),
	IsPublished: true,
}

type GuestbookCommentModel struct{}

func (m *GuestbookCommentModel) Insert(shortId uint64, guestbookId, parentId int64, authorName,
	authorEmail, authorSite, commentText, pageUrl string, isPublished bool) (int64, error) {
	return 2, nil
}

func (m *GuestbookCommentModel) Get(shortId uint64) (models.GuestbookComment, error) {
	switch shortId {
	case 1:
		return mockGuestbookComment, nil
	default:
		return models.GuestbookComment{}, models.ErrNoRecord
	}
}

func (m *GuestbookCommentModel) GetAll(guestbookId int64) ([]models.GuestbookComment, error) {
	switch guestbookId {
	case 1:
		return []models.GuestbookComment{mockGuestbookComment}, nil
	case 2:
		return []models.GuestbookComment{}, nil
	default:
		return []models.GuestbookComment{}, models.ErrNoRecord
	}
}

func (m *GuestbookCommentModel) GetDeleted(guestbookId int64) ([]models.GuestbookComment, error) {
	switch guestbookId {
	default:
		return []models.GuestbookComment{}, models.ErrNoRecord
	}
}

func (m *GuestbookCommentModel) GetUnpublished(guestbookId int64) ([]models.GuestbookComment, error) {
	switch guestbookId {
	default:
		return []models.GuestbookComment{}, models.ErrNoRecord
	}
}

func (m *GuestbookCommentModel) UpdateComment(comment *models.GuestbookComment) error {
	return nil
}
