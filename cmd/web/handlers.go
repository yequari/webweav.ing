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

func (app *application) home(w http.ResponseWriter, r *http.Request) {
    data := app.newCommonData(r)
    views.Home("Home", data).Render(r.Context(), w)
}

func (app *application) getUserRegister(w http.ResponseWriter, r *http.Request) {
    form := forms.UserRegistrationForm{}
    data := app.newCommonData(r)
    views.UserRegistration("User Registration", data, form).Render(r.Context(), w)
}

func (app *application) postUserRegister(w http.ResponseWriter, r *http.Request) {
    var form forms.UserRegistrationForm
    err := app.decodePostForm(r, &form)
    if err != nil {
        app.clientError(w, http.StatusBadRequest)
        return
    }
    form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
    form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
    form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
    form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
    form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")
    if !form.Valid() {
        data := app.newCommonData(r)
        w.WriteHeader(http.StatusUnprocessableEntity)
        views.UserRegistration("User Registration", data, form).Render(r.Context(), w)
        return
    }
    shortId := app.createShortId()
    err = app.users.Insert(shortId, form.Name, form.Email, form.Password)
    if err != nil {
        if errors.Is(err, models.ErrDuplicateEmail) {
            form.AddFieldError("email", "Email address is already in use")
            data := app.newCommonData(r)
            w.WriteHeader(http.StatusUnprocessableEntity)
            views.UserRegistration("User Registration", data, form).Render(r.Context(), w)
        } else {
            app.serverError(w, r, err)
        }
        return
    }
    app.sessionManager.Put(r.Context(), "flash", "Registration successful. Please log in.")
    http.Redirect(w, r, "/users/login", http.StatusSeeOther)
}

func (app *application) getUserLogin(w http.ResponseWriter, r *http.Request) {
    views.UserLogin("Login", app.newCommonData(r), forms.UserLoginForm{}).Render(r.Context(), w)
}

func (app *application) postUserLogin(w http.ResponseWriter, r *http.Request) {
    var form forms.UserLoginForm
    err := app.decodePostForm(r, &form)
    if err != nil {
        app.clientError(w, http.StatusBadRequest)
    }
    form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
    form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
    form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
    if !form.Valid() {
        data := app.newCommonData(r)
        w.WriteHeader(http.StatusUnprocessableEntity)
        views.UserLogin("Login", data, form).Render(r.Context(), w)
        return
    }
    id, err := app.users.Authenticate(form.Email, form.Password)
    if err != nil {
        if errors.Is(err, models.ErrInvalidCredentials) {
            form.AddNonFieldError("Email or password is incorrect")
            data := app.newCommonData(r)
            views.UserLogin("Login", data, form).Render(r.Context(), w)
        } else {
            app.serverError(w, r, err)
        }
        return
    }
    err = app.sessionManager.RenewToken(r.Context())
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    app.sessionManager.Put(r.Context(), "authenticatedUserId", id)
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) postUserLogout(w http.ResponseWriter, r *http.Request) {
    err := app.sessionManager.RenewToken(r.Context())
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    app.sessionManager.Remove(r.Context(), "authenticatedUserId")
    app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) getUsersList(w http.ResponseWriter, r *http.Request) {
    // skip templ conversion for this view, which will not be available in the final app
    // something similar will be available in the admin panel
    users, err := app.users.GetAll()
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    data := app.newTemplateData(r)
    data.Users = users
    app.render(w, r, http.StatusOK, "userlist.view.tmpl.html", data)
}

func (app *application) getUser(w http.ResponseWriter, r *http.Request) {
    slug := r.PathValue("id") 
    user, err := app.users.Get(slugToShortId(slug))
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            http.NotFound(w, r)
        } else {
            app.serverError(w, r, err)
        }
        return
    }
    data := app.newCommonData(r)
    views.UserProfile(user.Username, data, user).Render(r.Context(), w)
}

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

func (app *application) getCommentQueue(w http.ResponseWriter, r *http.Request) []models.GuestbookComment {
    guestbookSlug := r.PathValue("id")
    guestbook, err := app.guestbooks.Get(slugToShortId(guestbookSlug))
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            http.NotFound(w, r)
        } else {
            app.serverError(w, r, err)
        }
        return []models.GuestbookComment{}
    }

    comments, err := app.guestbookComments.GetQueue(guestbook.ID)
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            http.NotFound(w, r)
        } else {
            app.serverError(w, r, err)
        }
        return []models.GuestbookComment{}
    }

    return comments
}
