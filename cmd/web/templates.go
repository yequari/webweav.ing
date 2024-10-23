package main

import (
	"net/http"
	"path/filepath"
	"text/template"
	"time"

	"git.32bit.cafe/32bitcafe/guestbook/internal/models"
)

type templateData struct {
    CurrentYear int
    User models.User
    Users []models.User
    Guestbook models.Guestbook
    Guestbooks []models.Guestbook
    Comment models.GuestbookComment
    Comments []models.GuestbookComment
    Flash string
}

func humanDate(t time.Time) string {
    return t.Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap {
    "humanDate": humanDate,
    "encodeId": encodeIdB64,
    "decodeId": decodeIdB64,
}

func newTemplateCache() (map[string]*template.Template, error) {
    cache := map[string]*template.Template{}
    pages, err := filepath.Glob("./ui/html/pages/*.tmpl.html")
    if err != nil {
        return nil, err
    }
    for _, page := range pages {
        name := filepath.Base(page)
        ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.tmpl.html")
        if err != nil {
            return nil, err
        }
        ts, err = ts.ParseGlob("./ui/html/partials/*.tmpl.html")
        if err != nil {
            return nil, err
        }
        ts, err = ts.ParseFiles(page)
        if err != nil {
            return nil, err
        }
        cache[name] = ts
    }
    return cache, nil
}

func (app *application) newTemplateData(r *http.Request) templateData {
    return templateData {
        CurrentYear: time.Now().Year(),
        Flash: app.sessionManager.PopString(r.Context(), "flash"),
    }
}
