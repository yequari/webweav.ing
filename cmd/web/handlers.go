package main

import (
	"errors"
	"fmt"
	"net/http"

	"git.32bit.cafe/32bitcafe/guestbook/internal/models"
	"git.32bit.cafe/32bitcafe/guestbook/internal/validator"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
    data := app.newTemplateData(r)
    app.render(w, r, http.StatusOK, "home.tmpl.html", data)
}

type userRegistrationForm struct {
    Name        string  `schema:"username"`
    Email       string  `schema:"email"`
    Password    string  `schema:"password"`
    validator.Validator `schema:"-"`
}

func (app *application) getUserRegister(w http.ResponseWriter, r *http.Request) {
    data := app.newTemplateData(r)
    data.Form = userRegistrationForm{}
    app.render(w, r, http.StatusOK, "usercreate.view.tmpl.html", data)
}

func (app *application) postUserRegister(w http.ResponseWriter, r *http.Request) {
    var form userRegistrationForm
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
        data := app.newTemplateData(r)
        data.Form = form
        app.render(w, r, http.StatusUnprocessableEntity, "usercreate.view.tmpl.html", data)
        return
    }

    shortId := app.createShortId()
    err = app.users.Insert(shortId, form.Name, form.Email, form.Password)
    if err != nil {
        if errors.Is(err, models.ErrDuplicateEmail) {
            form.AddFieldError("email", "Email address is already in use")
            data := app.newTemplateData(r)
            data.Form = form
            app.render(w ,r, http.StatusUnprocessableEntity, "usercreate.view.tmpl.html", data)
        } else {
            app.serverError(w, r, err)
        }
        return
    }
    app.sessionManager.Put(r.Context(), "flash", "Registration successful. Please log in.")
    http.Redirect(w, r, "/users/login", http.StatusSeeOther)
}

type userLoginForm struct {
    Email       string  `schema:"email"`
    Password    string  `schema:"password"`
    validator.Validator `schema:"-"`
}

func (app *application) getUserLogin(w http.ResponseWriter, r *http.Request) {
    data := app.newTemplateData(r)
    data.Form = userLoginForm{}
    app.render(w, r, http.StatusOK, "login.view.tmpl.html", data)
}

func (app *application) postUserLogin(w http.ResponseWriter, r *http.Request) {
    var form userLoginForm

    err := app.decodePostForm(r, &form)
    if err != nil {
        app.clientError(w, http.StatusBadRequest)
    }

    form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
    form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
    form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

    if !form.Valid() {
        data := app.newTemplateData(r)
        data.Form = userLoginForm{}
        app.render(w, r, http.StatusUnprocessableEntity, "login.view.tmpl.html", data)
        return
    }

    id, err := app.users.Authenticate(form.Email, form.Password)
    if err != nil {
        if errors.Is(err, models.ErrInvalidCredentials) {
            form.AddNonFieldError("Email or password is incorrect")
            data := app.newTemplateData(r)
            data.Form = form
            app.render(w, r, http.StatusUnprocessableEntity, "login.view.tmpl.html", data)
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
    data := app.newTemplateData(r)
    data.User = user
    app.render(w, r, http.StatusOK, "user.view.tmpl.html", data)
}

func (app *application) getGuestbookCreate(w http.ResponseWriter, r* http.Request) {
    data := app.newTemplateData(r)
    app.render(w, r, http.StatusOK, "guestbookcreate.view.tmpl.html", data)
}

func (app *application) postGuestbookCreate(w http.ResponseWriter, r* http.Request) {
    err := r.ParseForm()
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    siteUrl := r.Form.Get("siteurl")
    shortId := app.createShortId()
    _, err = app.guestbooks.Insert(shortId, siteUrl, 0)
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    app.sessionManager.Put(r.Context(), "flash", "Guestbook successfully created!")
    http.Redirect(w, r, fmt.Sprintf("/guestbooks/%s", shortIdToSlug(shortId)), http.StatusSeeOther)
}

func (app *application) getGuestbookList(w http.ResponseWriter, r *http.Request) {
    userId := app.sessionManager.GetInt64(r.Context(), "authenticatedUserId")
    guestbooks, err := app.guestbooks.GetAll(userId)
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    user, err := app.users.GetById(userId)
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    data := app.newTemplateData(r)
    data.Guestbooks = guestbooks
    data.User = user
    app.render(w, r, http.StatusOK, "guestbooklist.view.tmpl.html", data)
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
    data := app.newTemplateData(r)
    data.Guestbook = guestbook
    data.Comments = comments
    app.render(w, r, http.StatusOK, "guestbook.view.tmpl.html", data)
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
    data := app.newTemplateData(r)
    data.Guestbook = guestbook
    data.Comments = comments
    app.render(w, r, http.StatusOK, "commentlist.view.tmpl.html", data)
}

type commentCreateForm struct {
    AuthorName  string  `schema:"authorname"`
    AuthorEmail string  `schema:"authoremail"`
    AuthorSite  string  `schema:"authorsite"`
    Content     string  `schema:"content,required"`
    validator.Validator `schema:"-"`
}

func (app *application) getGuestbookCommentCreate(w http.ResponseWriter, r *http.Request) {
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
    data.Form = commentCreateForm{}
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

    var form commentCreateForm
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
    app.sessionManager.Put(r.Context(), "flash", "Comment successfully posted!")
    http.Redirect(w, r, fmt.Sprintf("/guestbooks/%s", guestbookSlug), http.StatusSeeOther)
}
