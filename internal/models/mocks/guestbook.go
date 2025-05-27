package mocks

import (
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

type GuestbookModel struct{}

func (m *GuestbookModel) InitializeSettingsMap() error {
	return nil
}

func (m *GuestbookModel) Insert(shortId uint64, userId int64, websiteId int64, settings models.GuestbookSettings) (int64, error) {
	return 2, nil
}

func (m *GuestbookModel) Get(shortId uint64) (models.Guestbook, error) {
	switch shortId {
	case 1:
		return mockGuestbook, nil
	default:
		return models.Guestbook{}, models.ErrNoRecord
	}
}

func (m *GuestbookModel) GetAll(userId int64) ([]models.Guestbook, error) {
	switch userId {
	case 1:
		return []models.Guestbook{mockGuestbook}, nil
	default:
		return []models.Guestbook{}, models.ErrNoRecord
	}
}

func (m *GuestbookModel) UpdateGuestbookSettings(guestbookId int64, settings models.GuestbookSettings) error {
	return nil
}

func (m *GuestbookModel) UpdateSetting(guestbookId int64, setting models.Setting, value string) error {
	return nil
}

func (m *GuestbookModel) AddFilteredWord(guestbookId int64, word string) error {
	return nil
}

func (m *GuestbookModel) RemoveFilteredWord(guestbookId int64, word string) error {
	return nil
}
