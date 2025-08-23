package forms

import "git.32bit.cafe/32bitcafe/guestbook/internal/validator"

type UserRegistrationForm struct {
	Name                string `schema:"username"`
	Email               string `schema:"email"`
	Password            string `schema:"password"`
	validator.Validator `schema:"-"`
}

type UserLoginForm struct {
	Email               string `schema:"email"`
	Password            string `schema:"password"`
	validator.Validator `schema:"-"`
}

type CommentCreateForm struct {
	AuthorName          string `schema:"authorname"`
	AuthorEmail         string `schema:"authoremail"`
	AuthorSite          string `schema:"authorsite"`
	Content             string `schema:"content"`
	Redirect            string `schema:"redirect"`
	validator.Validator `schema:"-"`
}

type WebsiteCreateForm struct {
	Name                string `schema:"ws_name""`
	SiteUrl             string `schema:"ws_url"`
	AuthorName          string `schema:"ws_author"`
	validator.Validator `schema:"-"`
}

type WebsiteDeleteForm struct {
	Delete              string `schema:"delete"`
	validator.Validator `schema:"-"`
}

type UserSettingsForm struct {
	LocalTimezone       string `schema:"timezones"`
	validator.Validator `schema:"-"`
}

type WebsiteSettingsForm struct {
	SiteName            string `schema:"ws_name"`
	SiteUrl             string `schema:"ws_url"`
	AuthorName          string `schema:"ws_author"`
	Visibility          string `schema:"gb_visible"`
	CommentingEnabled   string `schema:"gb_commenting"`
	WidgetsEnabled      string `schema:"gb_remote"`
	validator.Validator `schema:"-"`
}

type AdminUserMgmtForm struct {
}

type InstallForm struct {
	Name                string `schema:"username"`
	Email               string `schema:"email"`
	Password            string `schema:"password"`
	validator.Validator `schema:"-"`
}
