package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
    mux := http.NewServeMux()
    fileServer := http.FileServer(http.Dir("./ui/static"))
    mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))
    
    dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)
    standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)

    mux.Handle("/{$}", dynamic.ThenFunc(app.home))
    mux.Handle("POST /guestbooks/{id}/comments/create", standard.ThenFunc(app.postGuestbookCommentCreate))
    mux.Handle("GET /guestbooks/{id}", dynamic.ThenFunc(app.getGuestbook))
    mux.Handle("GET /users/register", dynamic.ThenFunc(app.getUserRegister))
    mux.Handle("POST /users/register", dynamic.ThenFunc(app.postUserRegister))
    mux.Handle("GET /users/login", dynamic.ThenFunc(app.getUserLogin))
    mux.Handle("POST /users/login", dynamic.ThenFunc(app.postUserLogin))

    protected := dynamic.Append(app.requireAuthentication)

    mux.Handle("GET /users", protected.ThenFunc(app.getUsersList))
    mux.Handle("GET /users/{id}", protected.ThenFunc(app.getUser))
    mux.Handle("POST /users/logout", protected.ThenFunc(app.postUserLogout))
    mux.Handle("GET /guestbooks", protected.ThenFunc(app.getGuestbookList))
    mux.Handle("GET /guestbooks/create", protected.ThenFunc(app.getGuestbookCreate))
    mux.Handle("POST /guestbooks/create", protected.ThenFunc(app.postGuestbookCreate))
    mux.Handle("GET /guestbooks/{id}/comments/create", protected.ThenFunc(app.getGuestbookCommentCreate))


    return standard.Then(mux)
}

