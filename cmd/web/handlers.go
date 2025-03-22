package main

import (
	"net/http"

	"git.32bit.cafe/32bitcafe/guestbook/ui/views"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if app.isAuthenticated(r) {
		http.Redirect(w, r, "/websites", http.StatusSeeOther)
		return
	}
	views.Home("Home", app.newCommonData(r)).Render(r.Context(), w)
}
