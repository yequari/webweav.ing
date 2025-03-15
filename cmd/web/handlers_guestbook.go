package main


import (
	"errors"
	"fmt"
	"net/http"

        "git.32bit.cafe/32bitcafe/guestbook/internal/forms"
	"git.32bit.cafe/32bitcafe/guestbook/internal/models"
	"git.32bit.cafe/32bitcafe/guestbook/internal/validator"
	"git.32bit.cafe/32bitcafe/guestbook/ui/views"
)

func (app *application) getGuestbookCreate(w http.ResponseWriter, r* http.Request) {
    data := app.newCommonData(r)
    views.GuestbookCreate("New Guestbook", data).Render(r.Context(), w)
}

func (app *application) postGuestbookCreate(w http.ResponseWriter, r* http.Request) {
    userId := app.sessionManager.GetInt64(r.Context(), "authenticatedUserId")
    err := r.ParseForm()
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    siteUrl := r.Form.Get("siteurl")
    shortId := app.createShortId()
    _, err = app.guestbooks.Insert(shortId, siteUrl, userId)
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    app.sessionManager.Put(r.Context(), "flash", "Guestbook successfully created!")
    if r.Header.Get("HX-Request") == "true" {
        w.Header().Add("HX-Trigger", "newGuestbook")
        data := app.newTemplateData(r)
        app.renderHTMX(w, r, http.StatusOK, "guestbookcreatebutton.part.html", data)
        return
    }
    http.Redirect(w, r, fmt.Sprintf("/guestbooks/%s", shortIdToSlug(shortId)), http.StatusSeeOther)
}

func (app *application) getGuestbookList(w http.ResponseWriter, r *http.Request) {
    userId := app.sessionManager.GetInt64(r.Context(), "authenticatedUserId")
    guestbooks, err := app.guestbooks.GetAll(userId)
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    data := app.newCommonData(r)
    views.GuestbookList("Guestbooks", data, guestbooks).Render(r.Context(), w)
}

func (app *application) getGuestbook(w http.ResponseWriter, r *http.Request) {
    slug := r.PathValue("id")
    guestbook, err := app.guestbooks.Get(slugToShortId(slug))
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            http.NotFound(w, r)
        } else {
            app.serverError(w, r, err)
        }
        return
    }
    comments, err := app.guestbookComments.GetAll(guestbook.ID)
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    data := app.newCommonData(r)
    views.GuestbookView("Guestbook", data, guestbook, comments, forms.CommentCreateForm{}).Render(r.Context(), w)
}

func (app *application) getGuestbookDashboard(w http.ResponseWriter, r *http.Request) {
    slug := r.PathValue("id")
    guestbook, err := app.guestbooks.Get(slugToShortId(slug))
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            http.NotFound(w, r)
        } else {
            app.serverError(w, r, err)
        }
        return
    }
    comments, err := app.guestbookComments.GetAll(guestbook.ID)
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    data := app.newCommonData(r)
    views.GuestbookDashboardView("Guestbook", data, guestbook, comments).Render(r.Context(), w)
}

func (app *application) getGuestbookComments(w http.ResponseWriter, r *http.Request) {
    slug := r.PathValue("id")
    guestbook, err := app.guestbooks.Get(slugToShortId(slug))
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            http.NotFound(w, r)
        } else {
            app.serverError(w, r, err)
        }
        return
    }
    comments, err := app.guestbookComments.GetAll(guestbook.ID)
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    data := app.newCommonData(r)
    views.GuestbookDashboardCommentsView("Comments", data, guestbook, comments).Render(r.Context(), w)
}

func (app *application) getGuestbookCommentCreate(w http.ResponseWriter, r *http.Request) {
    // TODO: This will be the embeddable form
    slug := r.PathValue("id")
    guestbook, err := app.guestbooks.Get(slugToShortId(slug))
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            http.NotFound(w, r)
        } else {
            app.serverError(w, r, err)
        }
        return
    }
    data := app.newTemplateData(r)
    data.Guestbook = guestbook
    data.Form = forms.CommentCreateForm{}
    app.render(w, r, http.StatusOK, "commentcreate.view.tmpl.html", data)
}

func (app *application) postGuestbookCommentCreate(w http.ResponseWriter, r *http.Request) {
    guestbookSlug := r.PathValue("id")
    guestbook, err := app.guestbooks.Get(slugToShortId(guestbookSlug))
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            http.NotFound(w, r)
        } else {
            app.serverError(w, r, err)
        }
        return
    }

    var form forms.CommentCreateForm
    err = app.decodePostForm(r, &form)
    if err != nil {
        app.clientError(w, http.StatusBadRequest)
        return
    }

    form.CheckField(validator.NotBlank(form.AuthorName), "authorName", "This field cannot be blank")
    form.CheckField(validator.MaxChars(form.AuthorName, 256), "authorName", "This field cannot be more than 256 characters long")
    form.CheckField(validator.NotBlank(form.AuthorEmail), "authorEmail", "This field cannot be blank")
    form.CheckField(validator.MaxChars(form.AuthorEmail, 256), "authorEmail", "This field cannot be more than 256 characters long")
    form.CheckField(validator.NotBlank(form.AuthorSite), "authorSite", "This field cannot be blank")
    form.CheckField(validator.MaxChars(form.AuthorSite, 256), "authorSite", "This field cannot be more than 256 characters long")
    form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")

    if !form.Valid() {
        data := app.newTemplateData(r)
        data.Guestbook = guestbook
        data.Form = form
        app.render(w, r, http.StatusUnprocessableEntity, "commentcreate.view.tmpl.html", data)
        return
    }
    
    shortId := app.createShortId()
    _, err = app.guestbookComments.Insert(shortId, guestbook.ID, 0, form.AuthorName, form.AuthorEmail, form.AuthorSite, form.Content, "", true)
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    // app.sessionManager.Put(r.Context(), "flash", "Comment successfully posted!")
    http.Redirect(w, r, fmt.Sprintf("/guestbooks/%s", guestbookSlug), http.StatusSeeOther)
}

func (app *application) updateGuestbookComment(w http.ResponseWriter, r *http.Request) {
}

func (app *application) deleteGuestbookComment(w http.ResponseWriter, r *http.Request) {
    // slug := r.PathValue("id")
    // shortId := slugToShortId(slug)
    // app.guestbookComments.Delete(shortId)
}

func (app *application) getCommentQueue(w http.ResponseWriter, r *http.Request) {
    guestbookSlug := r.PathValue("id")
    guestbook, err := app.guestbooks.Get(slugToShortId(guestbookSlug))
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            http.NotFound(w, r)
        } else {
            app.serverError(w, r, err)
        }
        return
    }

    comments, err := app.guestbookComments.GetQueue(guestbook.ID)
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            http.NotFound(w, r)
        } else {
            app.serverError(w, r, err)
        }
        return
    }

    data := app.newCommonData(r)
    views.GuestbookDashboardCommentsView("Message Queue", data, guestbook, comments).Render(r.Context(), w)
}

func (app *application) putHideGuestbookComment(w http.ResponseWriter, r *http.Request) {

}

func (app *application) putDeleteGuestbookComment(w http.ResponseWriter, r *http.Request) {
}
