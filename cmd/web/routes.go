package main

import (
	"net/http"

	"git.32bit.cafe/32bitcafe/guestbook/ui"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	if app.config.environment == "PROD" {
		mux.Handle("GET /static/", http.FileServerFS(ui.Files))
	} else {
		fileServer := http.FileServer(http.Dir("./ui/static/"))
		mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))
	}

	mux.HandleFunc("GET /ping", ping)

	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)
	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)
	withCors := standard.Append(app.enableCors)

	mux.Handle("/{$}", dynamic.ThenFunc(app.home))
	mux.Handle("GET /websites/{id}/guestbook", dynamic.ThenFunc(app.getGuestbook))
	mux.Handle("GET /websites/{id}/guestbook/comments", withCors.ThenFunc(app.getGuestbookCommentsSerialized))
	mux.Handle("POST /websites/{id}/guestbook/comments/create/remote", standard.ThenFunc(app.postGuestbookCommentCreateRemote))
	mux.Handle("GET /websites/{id}/guestbook/comments/create", dynamic.ThenFunc(app.getGuestbookCommentCreate))
	mux.Handle("POST /websites/{id}/guestbook/comments/create", dynamic.ThenFunc(app.postGuestbookCommentCreate))
	mux.Handle("GET /users/register", dynamic.ThenFunc(app.getUserRegister))
	mux.Handle("POST /users/register", dynamic.ThenFunc(app.postUserRegister))
	mux.Handle("GET /users/login", dynamic.ThenFunc(app.getUserLogin))
	mux.Handle("POST /users/login", dynamic.ThenFunc(app.postUserLogin))
	mux.Handle("/users/login/oidc", dynamic.ThenFunc(app.userLoginOIDC))
	mux.Handle("/users/login/oidc/callback", dynamic.ThenFunc(app.userLoginOIDCCallback))
	mux.Handle("GET /help", dynamic.ThenFunc(app.notImplemented))
	mux.Handle("GET /about", dynamic.ThenFunc(app.about))

	protected := dynamic.Append(app.requireAuthentication)

	mux.Handle("GET /users/{id}", protected.ThenFunc(app.getUser))
	mux.Handle("POST /users/logout", protected.ThenFunc(app.postUserLogout))
	mux.Handle("GET /users/settings", protected.ThenFunc(app.getUserSettings))
	mux.Handle("PUT /users/settings", protected.ThenFunc(app.putUserSettings))
	mux.Handle("GET /guestbooks", protected.ThenFunc(app.getAllGuestbooks))
	mux.Handle("GET /websites", protected.ThenFunc(app.getWebsiteList))
	mux.Handle("GET /websites/create", protected.ThenFunc(app.getWebsiteCreate))
	mux.Handle("POST /websites/create", protected.ThenFunc(app.postWebsiteCreate))
	mux.Handle("GET /websites/{id}/dashboard", protected.ThenFunc(app.getWebsiteDashboard))
	mux.Handle("GET /websites/{id}/dashboard/guestbook/comments", protected.ThenFunc(app.getGuestbookComments))
	mux.Handle("GET /websites/{id}/dashboard/guestbook/comments/hidden", protected.ThenFunc(app.getCommentQueue))
	mux.Handle("DELETE /websites/{id}/dashboard/guestbook/comments/{commentId}", protected.ThenFunc(app.deleteGuestbookComment))
	mux.Handle("PUT /websites/{id}/dashboard/guestbook/comments/{commentId}", protected.ThenFunc(app.putHideGuestbookComment))
	mux.Handle("GET /websites/{id}/dashboard/settings", protected.ThenFunc(app.getWebsiteSettings))
	mux.Handle("PUT /websites/{id}/settings", protected.ThenFunc(app.putWebsiteSettings))
	mux.Handle("PUT /websites/{id}", protected.ThenFunc(app.deleteWebsite))
	mux.Handle("GET /websites/{id}/dashboard/guestbook/themes", protected.ThenFunc(app.getComingSoon))
	mux.Handle("GET /websites/{id}/dashboard/guestbook/customize", protected.ThenFunc(app.getComingSoon))

	adminOnly := protected.Append(app.requireAdmin)
	mux.Handle("GET /admin", adminOnly.ThenFunc(app.getAdminPanelLanding))
	mux.Handle("GET /admin/users", adminOnly.ThenFunc(app.getAdminPanelAllUsers))
	mux.Handle("GET /admin/users/{id}", adminOnly.ThenFunc(app.getAdminPanelUser))
	mux.Handle("GET /admin/users/{id}/edit", adminOnly.ThenFunc(app.getAdminPanelUserMgmtForm))
	mux.Handle("GET /admin/users/{id}/detail", adminOnly.ThenFunc(app.getAdminPanelUserMgmtDetail))
	mux.Handle("PUT /admin/users/{id}/edit", adminOnly.ThenFunc(app.putAdminPanelUserMgmtForm))
	mux.Handle("PUT /admin/users/{id}/ban", adminOnly.ThenFunc(app.putAdminPanelBanUser))
	mux.Handle("PUT /admin/users/{id}/unban", adminOnly.ThenFunc(app.putAdminPanelUnbanUser))
	mux.Handle("GET /admin/websites", adminOnly.ThenFunc(app.getAdminPanelWebsites))
	mux.Handle("GET /admin/websites/{id}", adminOnly.ThenFunc(app.getAdminPanelWebsiteDetails))

	return standard.Then(mux)
}
