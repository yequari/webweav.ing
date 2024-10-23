package main

import (
    "net/http"
)

func (app *application) routes() http.Handler {
    mux := http.NewServeMux()
    fileServer := http.FileServer(http.Dir("./ui/static"))
    mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

    mux.Handle("/{$}", app.sessionManager.LoadAndSave(http.HandlerFunc(app.home)))
    mux.Handle("GET /users", app.sessionManager.LoadAndSave(http.HandlerFunc(app.getUsersList)))
    mux.Handle("GET /users/{id}", app.sessionManager.LoadAndSave(http.HandlerFunc(app.getUser)))
    mux.Handle("GET /users/register", app.sessionManager.LoadAndSave(http.HandlerFunc(app.getUserRegister)))
    mux.Handle("POST /users/register", app.sessionManager.LoadAndSave(http.HandlerFunc(app.postUserRegister)))
    mux.Handle("GET /guestbooks", app.sessionManager.LoadAndSave(http.HandlerFunc(app.getGuestbookList)))
    mux.Handle("GET /guestbooks/{id}", app.sessionManager.LoadAndSave(http.HandlerFunc(app.getGuestbook)))
    mux.Handle("GET /guestbooks/create", app.sessionManager.LoadAndSave(http.HandlerFunc(app.getGuestbookCreate)))
    mux.Handle("POST /guestbooks/create", app.sessionManager.LoadAndSave(http.HandlerFunc(app.postGuestbookCreate)))
    mux.Handle("GET /guestbooks/{id}/comments/create", app.sessionManager.LoadAndSave(http.HandlerFunc(app.getGuestbookCommentCreate)))
    mux.Handle("POST /guestbooks/{id}/comments/create", app.sessionManager.LoadAndSave(http.HandlerFunc(app.postGuestbookCommentCreate)))
    return app.recoverPanic(app.logRequest(commonHeaders(mux)))
}

