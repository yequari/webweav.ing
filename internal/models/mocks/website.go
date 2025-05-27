package mocks

import (
	"time"

	"git.32bit.cafe/32bitcafe/guestbook/internal/models"
)

var mockWebsite = models.Website{
	ID:         1,
	ShortId:    1,
	Name:       "Example",
	SiteUrl:    "example.com",
	AuthorName: "John Test",
	UserId:     1,
	Created:    time.Now(),
	Guestbook:  mockGuestbook,
}

type WebsiteModel struct{}

func (m *WebsiteModel) Insert(shortId uint64, userId int64, siteName, siteUrl, authorName string) (int64, error) {
	return 2, nil
}

func (m *WebsiteModel) Get(shortId uint64) (models.Website, error) {
	switch shortId {
	case 1:
		return mockWebsite, nil
	default:
		return models.Website{}, models.ErrNoRecord
	}
}

func (m *WebsiteModel) GetAllUser(userId int64) ([]models.Website, error) {
	return []models.Website{mockWebsite}, nil
}

func (m *WebsiteModel) GetById(id int64) (models.Website, error) {
	switch id {
	case 1:
		return mockWebsite, nil
	default:
		return models.Website{}, models.ErrNoRecord
	}
}

func (m *WebsiteModel) GetAll() ([]models.Website, error) {
	return []models.Website{mockWebsite}, nil
}
