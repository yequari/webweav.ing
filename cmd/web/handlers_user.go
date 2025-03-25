package main

import (
	"errors"
	"net/http"

	"git.32bit.cafe/32bitcafe/guestbook/internal/forms"
	"git.32bit.cafe/32bitcafe/guestbook/internal/models"
	"git.32bit.cafe/32bitcafe/guestbook/internal/validator"
	"git.32bit.cafe/32bitcafe/guestbook/ui/views"
)

func (app *application) getUserRegister(w http.ResponseWriter, r *http.Request) {
	form := forms.UserRegistrationForm{}
	data := app.newCommonData(r)
	views.UserRegistration("User Registration", data, form).Render(r.Context(), w)
}

func (app *application) getUserLogin(w http.ResponseWriter, r *http.Request) {
	views.UserLogin("Login", app.newCommonData(r), forms.UserLoginForm{}).Render(r.Context(), w)
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
	w.Header().Add("HX-Redirect", "/")
	// http.Redirect(w, r, "/", http.StatusSeeOther)
}

// func (app *application) getUsersList(w http.ResponseWriter, r *http.Request) {
//     // skip templ conversion for this view, which will not be available in the final app
//     // something similar will be available in the admin panel
//     users, err := app.users.GetAll()
//     if err != nil {
//         app.serverError(w, r, err)
//         return
//     }
//     data := app.newTemplateData(r)
//     data.Users = users
//     app.render(w, r, http.StatusOK, "userlist.view.tmpl.html", data)
// }

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
