package mocks

import (
	"net/url"
	"time"

	"git.32bit.cafe/32bitcafe/guestbook/internal/models"
)

var mockGuestbook = models.Guestbook{
	ID:        1,
	ShortId:   1,
	UserId:    1,
	WebsiteId: 1,
	Created:   time.Now(),
	IsActive:  true,
	Settings: models.GuestbookSettings{
		IsCommentingEnabled:   true,
		IsVisible:             true,
		FilteredWords:         make([]string, 0),
		AllowRemoteHostAccess: true,
	},
}

var mockWebsite = models.Website{
	ID:      1,
	ShortId: 1,
	Name:    "Example",
	// SiteUrl:    "example.com",
	Url: &url.URL{
		Scheme: "http",
		Host:   "example.com",
	},
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

func (m *WebsiteModel) Update(w models.Website) error {
	return nil
}

func (m *WebsiteModel) InitializeSettingsMap() error {
	return nil
}

func (m *WebsiteModel) UpdateGuestbookSettings(guestbookId int64, settings models.GuestbookSettings) error {
	return nil
}

func (m *WebsiteModel) UpdateSetting(guestbookId int64, setting models.Setting, value string) error {
	return nil
}
