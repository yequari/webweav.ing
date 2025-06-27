package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"git.32bit.cafe/32bitcafe/guestbook/internal/forms"
	"git.32bit.cafe/32bitcafe/guestbook/internal/models"
	"git.32bit.cafe/32bitcafe/guestbook/internal/validator"
	"git.32bit.cafe/32bitcafe/guestbook/ui/views"
)

func (app *application) getWebsiteCreate(w http.ResponseWriter, r *http.Request) {
	form := forms.WebsiteCreateForm{}
	data := app.newCommonData(r)
	views.WebsiteCreate("Add Website", data, form).Render(r.Context(), w)
}

func (app *application) postWebsiteCreate(w http.ResponseWriter, r *http.Request) {
	userId := app.sessionManager.GetInt64(r.Context(), "authenticatedUserId")
	var form forms.WebsiteCreateForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.AuthorName), "authorName", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.AuthorName, 256), "authorName", "This field cannot exceed 256 characters")
	form.CheckField(validator.NotBlank(form.Name), "sitename", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Name, 256), "sitename", "This field cannot exceed 256 characters")
	form.CheckField(validator.NotBlank(form.SiteUrl), "siteurl", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.SiteUrl, 512), "siteurl", "This field cannot exceed 512 characters")

	u, err := url.Parse(form.SiteUrl)

	if !form.Valid() || err != nil {
		data := app.newCommonData(r)
		w.WriteHeader(http.StatusUnprocessableEntity)
		views.WebsiteCreate("Add a Website", data, form).Render(r.Context(), w)
	}
	websiteShortID := app.createShortId()
	_, err = app.websites.Insert(websiteShortID, userId, form.Name, u.Host, form.AuthorName)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	app.sessionManager.Put(r.Context(), "flash", "Website successfully registered!")
	http.Redirect(w, r, fmt.Sprintf("/websites/%s/dashboard", shortIdToSlug(websiteShortID)), http.StatusSeeOther)
}

func (app *application) getWebsiteDashboard(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	user := app.getCurrentUser(r)
	website, err := app.websites.Get(slugToShortId(slug))
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	if user.ID != website.UserId {
		app.clientError(w, http.StatusUnauthorized)
	}
	data := app.newCommonData(r)
	views.WebsiteDashboard("Guestbook", data, website).Render(r.Context(), w)
}

func (app *application) getWebsiteList(w http.ResponseWriter, r *http.Request) {

	userId := app.sessionManager.GetInt64(r.Context(), "authenticatedUserId")
	websites, err := app.websites.GetAllUser(userId)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if r.Header.Get("HX-Request") == "true" {
		views.HxWebsiteList(websites)
		return
	}
	data := app.newCommonData(r)
	views.WebsiteList("My Websites", data, websites).Render(r.Context(), w)
}

func (app *application) getComingSoon(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	user := app.getCurrentUser(r)
	website, err := app.websites.Get(slugToShortId(slug))
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	if website.UserId != user.ID {
		app.clientError(w, http.StatusForbidden)
	}
	views.WebsiteDashboardComingSoon("Coming Soon", app.newCommonData(r), website).Render(r.Context(), w)
}
