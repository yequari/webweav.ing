package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"git.32bit.cafe/32bitcafe/guestbook/internal/models"
	"github.com/google/uuid"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
    app.render(w, r, http.StatusOK, "home.tmpl.html", templateData{})
}

func (app *application) getUserRegister(w http.ResponseWriter, r *http.Request) {
    app.render(w, r, http.StatusOK, "usercreate.view.tmpl.html", templateData{})
}

func (app *application) postUserRegister(w http.ResponseWriter, r *http.Request) {
    err := r.ParseForm()
    if err != nil {
        app.serverError(w, r, err)
    }
    username := r.Form.Get("username")
    email := r.Form.Get("email")
    rawid, err := app.users.Insert(username, email)
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    id, err := encodeIdB64(rawid)
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    http.Redirect(w, r, fmt.Sprintf("/users/%s", id), http.StatusSeeOther)
}

func (app *application) getUsersList(w http.ResponseWriter, r *http.Request) {
    users, err := app.users.GetAll()
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    app.render(w, r, http.StatusOK, "userlist.view.tmpl.html", templateData{
        Users: users,
    })
}

func (app *application) getUser(w http.ResponseWriter, r *http.Request) {
    rawid := r.PathValue("id") 
    id, err := decodeIdB64(rawid)
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    user, err := app.users.Get(id)
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            http.NotFound(w, r)
        } else {
            app.serverError(w, r, err)
        }
        return
    }
    app.render(w, r, http.StatusOK, "user.view.tmpl.html", templateData{
        User: user,
    })
}

func (app *application) getGuestbookCreate(w http.ResponseWriter, r* http.Request) {
    app.render(w, r, http.StatusOK, "guestbookcreate.view.tmpl.html", templateData{})
}

func (app *application) postGuestbookCreate(w http.ResponseWriter, r* http.Request) {
    err := r.ParseForm()
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    siteUrl := r.Form.Get("siteurl")
    app.logger.Debug("creating guestbook for site", "siteurl", siteUrl)
    userId := getUserId()
    rawid, err := app.guestbooks.Insert(siteUrl, userId)
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    id, err := encodeIdB64(rawid)
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    app.sessionManager.Put(r.Context(), "flash", "Guestbook successfully created!")
    http.Redirect(w, r, fmt.Sprintf("/guestbooks/%s", id), http.StatusSeeOther)
}

func (app *application) getGuestbookList(w http.ResponseWriter, r *http.Request) {
    userId := getUserId()
    guestbooks, err := app.guestbooks.GetAll(userId)
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    app.render(w, r, http.StatusOK, "guestbooklist.view.tmpl.html", templateData{
        Guestbooks: guestbooks,
    })
}

func (app *application) getGuestbook(w http.ResponseWriter, r *http.Request) {
    rawId := r.PathValue("id")
    id, err := decodeIdB64(rawId)
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    guestbook, err := app.guestbooks.Get(id)
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            http.NotFound(w, r)
        } else {
            app.serverError(w, r, err)
        }
        return
    }
    comments, err := app.guestbookComments.GetAll(id)
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    data := app.newTemplateData(r)
    data.Guestbook = guestbook
    data.Comments = comments
    app.render(w, r, http.StatusOK, "guestbook.view.tmpl.html", data)
}

func (app *application) getGuestbookComments(w http.ResponseWriter, r *http.Request) {
    rawId := r.PathValue("id")
    id, err := decodeIdB64(rawId)
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    guestbook, err := app.guestbooks.Get(id)
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            http.NotFound(w, r)
        } else {
            app.serverError(w, r, err)
        }
        return
    }
    comments, err := app.guestbookComments.GetAll(id)
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    app.render(w, r, http.StatusOK, "commentlist.view.tmpl.html", templateData{
        Guestbook: guestbook,
        Comments: comments,
    })
}

func (app *application) getGuestbookCommentCreate(w http.ResponseWriter, r *http.Request) {
    rawId := r.PathValue("id")
    id, err := decodeIdB64(rawId)
    if err != nil {
        http.NotFound(w, r)
    }
    app.render(w, r, http.StatusOK, "commentcreate.view.tmpl.html", templateData{
        Guestbook: models.Guestbook{
            ID: id,
        },
    })
}

func (app *application) postGuestbookCommentCreate(w http.ResponseWriter, r *http.Request) {
    rawGbId := r.PathValue("id")
    gbId, err := decodeIdB64(rawGbId)
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    err = r.ParseForm()
    if err != nil {
        app.clientError(w, http.StatusBadRequest)
        return
    }
    authorName := r.PostForm.Get("authorname")
    authorEmail := r.PostForm.Get("authoremail")
    authorSite := r.PostForm.Get("authorsite")
    content := r.PostForm.Get("content")

    fieldErrors := make(map[string]string)
    if strings.TrimSpace(authorName) == "" {
        fieldErrors["title"] = "This field cannot be blank"
    } else if utf8.RuneCountInString(authorName) > 256 {
        fieldErrors["title"] = "This field cannot be more than 256 characters long"
    }
    if strings.TrimSpace(content) == "" {
        fieldErrors["content"] = "This field cannot be blank"
    }
    if len(fieldErrors) > 0 {
        fmt.Fprint(w, fieldErrors)
        return
    }
    
    commentId, err := app.guestbookComments.Insert(gbId, uuid.UUID{}, authorName, authorEmail, authorSite, content, "", true)
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    _, err = encodeIdB64(commentId)
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    http.Redirect(w, r, fmt.Sprintf("/guestbooks/%s", rawGbId), http.StatusSeeOther)
}
