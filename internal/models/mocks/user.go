package mocks

import (
	"errors"
	"time"

	"git.32bit.cafe/32bitcafe/guestbook/internal/models"
)

var mockUser = models.User{
	ID:       1,
	ShortId:  1,
	Username: "tester",
	Email:    "test@example.com",
	Deleted:  false,
	IsBanned: false,
	Created:  time.Now(),
	Settings: mockUserSettings,
}

var mockUserSettings = models.UserSettings{
	LocalTimezone: time.UTC,
}

type UserModel struct {
	Settings map[string]models.Setting
}

func (m *UserModel) InitializeSettingsMap() error {
	return nil
}

func (m *UserModel) Insert(shortId uint64, username string, email string, password string, settings models.UserSettings) error {
	switch email {
	case "dupe@example.com":
		return models.ErrDuplicateEmail
	default:
		return nil
	}
}

func (m *UserModel) InsertWithoutPassword(shortId uint64, username string, email string, password string, settings models.UserSettings) (int64, error) {
	switch email {
	case "dupe@example.com":
		return -1, models.ErrDuplicateEmail
	default:
		return 2, nil
	}
}

func (m *UserModel) Get(shortId uint64) (models.User, error) {
	switch shortId {
	case 1:
		return mockUser, nil
	default:
		return models.User{}, models.ErrNoRecord
	}
}

func (m *UserModel) GetById(id int64) (models.User, error) {
	switch id {
	case 1:
		return mockUser, nil
	default:
		return models.User{}, models.ErrNoRecord
	}
}

func (m *UserModel) GetAll() ([]models.User, error) {
	return []models.User{mockUser}, nil
}

func (m *UserModel) Authenticate(email, password string) (int64, error) {
	if email == "test@example.com" && password == "password" {
		return 1, nil
	}
	return 0, models.ErrInvalidCredentials
}

func (m *UserModel) Exists(id int64) (bool, error) {
	switch id {
	case 1:
		return true, nil
	default:
		return false, nil
	}
}

func (m *UserModel) GetSettings(userId int64) (models.UserSettings, error) {
	return mockUserSettings, nil
}

func (m *UserModel) UpdateUserSettings(userId int64, settings models.UserSettings) error {
	return nil
}

func (m *UserModel) UpdateSetting(userId int64, setting models.Setting, value string) error {
	return nil
}

func (m *UserModel) GetBySubject(subject string) (int64, error) {
	if subject == "goodSubject" {
		return 1, nil
	} else if subject == "foo" {
		return -1, models.ErrNoRecord
	}
	return -1, errors.New("Unexpected Error")
}

func (m *UserModel) GetByEmail(email string) (int64, error) {
	if email == "test@example.com" {
		return 1, nil
	}
	return -1, models.ErrNoRecord
}

func (m *UserModel) UpdateSubject(userId int64, subject string) error {
	if userId == 1 {
		return nil
	}
	return errors.New("invalid")
}
