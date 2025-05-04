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
	Content             string `schema:"content,required"`
	validator.Validator `schema:"-"`
}

type WebsiteCreateForm struct {
	Name                string `schema:"sitename"`
	SiteUrl             string `schema:"siteurl"`
	AuthorName          string `schema:"authorname"`
	validator.Validator `schema:"-"`
}

type UserSettingsForm struct {
	LocalTimezone       string `schema:"timezones"`
	validator.Validator `schema:"-"`
}
