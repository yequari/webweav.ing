package main

import (
    "net/http"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
    app.render(w, r, http.StatusOK, "home.tmpl.html", templateData{})
}

func getUserRegister(w http.ResponseWriter, r *http.Request) {
}

func postUserRegister(w http.ResponseWriter, r *http.Request) {
}

func getUsersList(w http.ResponseWriter, r *http.Request) {
}

func getUser(w http.ResponseWriter, r *http.Request) {
}

func postGuestbooksCreate(w http.ResponseWriter, r* http.Request) {
}

func getGuestbooksList(w http.ResponseWriter, r *http.Request) {
}

func getGuestbook(w http.ResponseWriter, r *http.Request) {
}
