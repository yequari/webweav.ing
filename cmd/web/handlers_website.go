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
	form.CheckField(validator.Matches(form.SiteUrl, validator.WebRX), "siteurl", "This field must be a valid URL (including http:// or https://)")

	u, err := url.Parse(form.SiteUrl)
	if err != nil {
		form.CheckField(false, "siteurl", "This field must be a valid URL")
	}
	if !form.Valid() {
		data := app.newCommonData(r)
		w.WriteHeader(http.StatusUnprocessableEntity)
		views.WebsiteCreate("Add a Website", data, form).Render(r.Context(), w)
		return
	}
	websiteShortID := app.createShortId()
	_, err = app.websites.Insert(websiteShortID, userId, form.Name, u.String(), form.AuthorName)
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
	views.WebsiteDashboard(fmt.Sprintf("%s - Dashboard", website.Name), data, website).Render(r.Context(), w)
}

func (app *application) getWebsiteList(w http.ResponseWriter, r *http.Request) {

	userId := app.sessionManager.GetInt64(r.Context(), "authenticatedUserId")
	websites, err := app.websites.GetAllUser(userId)
	if err != nil {
		app.serverError(w, r, err)
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

func (app *application) getWebsiteSettings(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	website, err := app.websites.Get(slugToShortId(slug))
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	var form forms.WebsiteSettingsForm
	data := app.newCommonData(r)
	views.WebsiteDashboardSettings(data, website, form).Render(r.Context(), w)
}

func (app *application) putWebsiteSettings(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	website, err := app.websites.Get(slugToShortId(slug))
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	var form forms.WebsiteSettingsForm
	err = app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		app.serverError(w, r, err)
		return
	}
	form.CheckField(validator.NotBlank(form.AuthorName), "ws_author", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.AuthorName, 256), "ws_author", "This field cannot exceed 256 characters")
	form.CheckField(validator.NotBlank(form.SiteName), "ws_name", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.SiteName, 256), "ws_name", "This field cannot exceed 256 characters")
	form.CheckField(validator.NotBlank(form.SiteUrl), "ws_url", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.SiteUrl, 512), "ws_url", "This field cannot exceed 512 characters")
	form.CheckField(validator.Matches(form.SiteUrl, validator.WebRX), "ws_url", "This field must be a valid URL (including http:// or https://)")
	form.CheckField(validator.PermittedValue(form.Visibility, "true", "false"), "gb_visible", "Invalid value")
	form.CheckField(validator.PermittedValue(form.CommentingEnabled, models.ValidDisableDurations...), "gb_visible", "Invalid value")
	form.CheckField(validator.PermittedValue(form.WidgetsEnabled, "true", "false"), "gb_remote", "Invalid value")
	if !form.Valid() {
		data := app.newCommonData(r)
		views.SettingsForm(data, website, form, "").Render(r.Context(), w)
		return
	}

	gbSettings, err := convertFormToGuestbookSettings(form)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = app.websites.UpdateGuestbookSettings(website.Guestbook.ID, gbSettings)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	u, err := url.Parse(form.SiteUrl)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	website.Name = form.SiteName
	website.AuthorName = form.AuthorName
	website.Url = u

	err = app.websites.Update(website)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Settings changed successfully")
	data := app.newCommonData(r)
	views.SettingsForm(data, website, forms.WebsiteSettingsForm{}, "Settings changed successfully").Render(r.Context(), w)

}

func (app *application) deleteWebsite(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	website, err := app.websites.Get(slugToShortId(slug))
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	var form forms.WebsiteDeleteForm
	err = app.decodePostForm(r, &form)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	form.CheckField(validator.Equals(website.Name, form.Delete), "delete", "Input must match site name exactly")
	if !form.Valid() {
		data := app.newCommonData(r)
		views.DeleteForm(data, website, form).Render(r.Context(), w)
		return
	}

	err = app.websites.Delete(website.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Website Deleted")
	w.Header().Add("HX-Redirect", "/websites")
	w.WriteHeader(http.StatusOK)
}
