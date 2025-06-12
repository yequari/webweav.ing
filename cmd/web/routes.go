package main

import (
	"net/http"

	"git.32bit.cafe/32bitcafe/guestbook/ui"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.FileServerFS(ui.Files))

	mux.HandleFunc("GET /ping", ping)

	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)
	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)
	withCors := standard.Append(app.enableCors)

	mux.Handle("/{$}", dynamic.ThenFunc(app.home))
	mux.Handle("GET /websites/{id}/guestbook", dynamic.ThenFunc(app.getGuestbook))
	mux.Handle("GET /websites/{id}/guestbook/comments", withCors.ThenFunc(app.getGuestbookCommentsSerialized))
	mux.Handle("GET /websites/{id}/guestbook/comments/create", dynamic.ThenFunc(app.getGuestbookCommentCreate))
	mux.Handle("POST /websites/{id}/guestbook/comments/create", dynamic.ThenFunc(app.postGuestbookCommentCreate))
	mux.Handle("GET /users/register", dynamic.ThenFunc(app.getUserRegister))
	mux.Handle("POST /users/register", dynamic.ThenFunc(app.postUserRegister))
	mux.Handle("GET /users/login", dynamic.ThenFunc(app.getUserLogin))
	mux.Handle("POST /users/login", dynamic.ThenFunc(app.postUserLogin))
	mux.Handle("GET /help", dynamic.ThenFunc(app.notImplemented))

	protected := dynamic.Append(app.requireAuthentication)

	// mux.Handle("GET /users", protected.ThenFunc(app.getUsersList))
	mux.Handle("GET /users/{id}", protected.ThenFunc(app.getUser))
	mux.Handle("POST /users/logout", protected.ThenFunc(app.postUserLogout))
	mux.Handle("GET /users/settings", protected.ThenFunc(app.getUserSettings))
	mux.Handle("PUT /users/settings", protected.ThenFunc(app.putUserSettings))
	mux.Handle("GET /users/privacy", protected.ThenFunc(app.notImplemented))
	mux.Handle("GET /guestbooks", protected.ThenFunc(app.getAllGuestbooks))

	mux.Handle("GET /websites", protected.ThenFunc(app.getWebsiteList))
	mux.Handle("GET /websites/create", protected.ThenFunc(app.getWebsiteCreate))
	mux.Handle("POST /websites/create", protected.ThenFunc(app.postWebsiteCreate))
	mux.Handle("GET /websites/{id}/dashboard", protected.ThenFunc(app.getWebsiteDashboard))
	mux.Handle("GET /websites/{id}/dashboard/guestbook/comments", protected.ThenFunc(app.getGuestbookComments))
	mux.Handle("GET /websites/{id}/dashboard/guestbook/comments/queue", protected.ThenFunc(app.getCommentQueue))
	mux.Handle("DELETE /websites/{id}/dashboard/guestbook/comments/{commentId}", protected.ThenFunc(app.deleteGuestbookComment))
	mux.Handle("PUT /websites/{id}/dashboard/guestbook/comments/{commentId}", protected.ThenFunc(app.putHideGuestbookComment))
	mux.Handle("GET /websites/{id}/dashboard/guestbook/settings", protected.ThenFunc(app.getGuestbookSettings))
	mux.Handle("PUT /websites/{id}/dashboard/guestbook/settings", protected.ThenFunc(app.putGuestbookSettings))
	mux.Handle("GET /websites/{id}/dashboard/guestbook/comments/trash", protected.ThenFunc(app.getCommentTrash))
	mux.Handle("GET /websites/{id}/dashboard/guestbook/themes", protected.ThenFunc(app.getComingSoon))
	mux.Handle("GET /websites/{id}/dashboard/guestbook/customize", protected.ThenFunc(app.getComingSoon))

	return standard.Then(mux)
}
