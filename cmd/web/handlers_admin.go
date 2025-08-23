package main

import (
	"errors"
	"fmt"
	"net/http"

	"git.32bit.cafe/32bitcafe/guestbook/internal/models"
	"git.32bit.cafe/32bitcafe/guestbook/ui/views"
)

func (app *application) getAdminPanelLanding(w http.ResponseWriter, r *http.Request) {
	data := app.newCommonData(r)
	views.AdminPanelLandingView("Admin Panel", data).Render(r.Context(), w)
}

func (app *application) getAdminPanelAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := app.users.GetAll()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newCommonData(r)
	views.AdminPanelUsersView("All Users - Admin", data, users).Render(r.Context(), w)
}

func (app *application) getAdminPanelUser(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	u, err := app.users.Get(slugToShortId(slug))
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	data := app.newCommonData(r)
	views.AdminPanelUserMgmtView(fmt.Sprintf("User Management - %s", u.Username), data, u).Render(r.Context(), w)
}

func (app *application) putAdminPanelBanUser(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	u, err := app.users.Get(slugToShortId(slug))
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	err = app.users.BanUser(u.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}

func (app *application) getAdminPanelWebsites(w http.ResponseWriter, r *http.Request) {
}
