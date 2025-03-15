package main

import (
	"net/http"
	"git.32bit.cafe/32bitcafe/guestbook/ui/views"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
    data := app.newCommonData(r)
    views.Home("Home", data).Render(r.Context(), w)
}
