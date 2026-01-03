package main

import (
	"errors"
	"net/http"
	"time"

	"git.32bit.cafe/32bitcafe/guestbook/internal/forms"
	"git.32bit.cafe/32bitcafe/guestbook/internal/models"
	"git.32bit.cafe/32bitcafe/guestbook/internal/validator"
	"git.32bit.cafe/32bitcafe/guestbook/ui/views"
	"github.com/coreos/go-oidc/v3/oidc"
)

func (app *application) getUserRegister(w http.ResponseWriter, r *http.Request) {
	if !app.config.localAuthEnabled {
		http.Redirect(w, r, "/users/login/oidc", http.StatusFound)
	}
	form := forms.UserRegistrationForm{}
	data := app.newCommonData(r)
	views.UserRegistration("User Registration", data, form).Render(r.Context(), w)
}

func (app *application) getUserLogin(w http.ResponseWriter, r *http.Request) {
	if !app.config.localAuthEnabled {
		http.Redirect(w, r, "/users/login/oidc", http.StatusFound)
	}
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
	settings := DefaultUserSettings()
	err = app.users.Insert(shortId, form.Name, form.Email, form.Password, settings)
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

func (app *application) userLoginOIDC(w http.ResponseWriter, r *http.Request) {
	if !app.config.oauthEnabled {
		http.Redirect(w, r, "/users/login", http.StatusFound)
	}
	state, err := randString(16)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	nonce, err := randString(16)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	setCallbackCookie(w, r, "state", state)
	setCallbackCookie(w, r, "nonce", nonce)

	http.Redirect(w, r, app.config.oauth.config.AuthCodeURL(state, oidc.Nonce(nonce)), http.StatusFound)
}

func (app *application) userLoginOIDCCallback(w http.ResponseWriter, r *http.Request) {
	state, err := r.Cookie("state")
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if r.URL.Query().Get("state") != state.Value {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	oauth2Token, err := app.config.oauth.config.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		app.logger.Error("Failed to exchange token")
		app.serverError(w, r, err)
		return
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		app.serverError(w, r, errors.New("No id_token field in oauth2 token"))
		return
	}
	idToken, err := app.config.oauth.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		app.logger.Error("Failed to verify ID token")
		app.serverError(w, r, err)
		return
	}

	nonce, err := r.Cookie("nonce")
	if err != nil {
		app.logger.Error("nonce not found")
		app.serverError(w, r, err)
		return
	}
	if idToken.Nonce != nonce.Value {
		app.serverError(w, r, errors.New("nonce did not match"))
		return
	}

	oauth2Token.AccessToken = "*REDACTED*"

	var t models.UserIdToken
	if err := idToken.Claims(&t); err != nil {
		app.serverError(w, r, err)
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// search for user by subject
	id, err := app.users.GetBySubject(t.Subject)
	if err != nil && errors.Is(err, models.ErrNoRecord) {
		// if no user is found, check if they have signed up by email already
		id, err = app.users.GetByEmail(t.Email)
		if err == nil {
			// if user is found by email, update subject to match them in the first step next time
			err2 := app.users.UpdateSubject(id, t.Subject)
			if err2 != nil {
				app.serverError(w, r, err2)
				return
			}
		}
	} else if err != nil {
		app.serverError(w, r, err)
		return
	}
	if err != nil && errors.Is(err, models.ErrNoRecord) {
		// if no user is found by subject or email, create a new user
		id, err = app.users.InsertWithoutPassword(app.createShortId(), t.Username, t.Email, t.Subject, DefaultUserSettings())
		if err != nil {
			app.serverError(w, r, err)
		}
	} else if err != nil {
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

func (app *application) getUserSettings(w http.ResponseWriter, r *http.Request) {
	data := app.newCommonData(r)
	views.UserSettingsView(data, app.timezones).Render(r.Context(), w)
}

func (app *application) putUserSettings(w http.ResponseWriter, r *http.Request) {
	user := app.getCurrentUser(r)
	var form forms.UserSettingsForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		app.serverError(w, r, err)
		return
	}
	form.CheckField(validator.PermittedValue(form.LocalTimezone, app.timezones...), "timezone", "Invalid value")
	if !form.Valid() {
		// TODO: rerender template with errors
		app.clientError(w, http.StatusUnprocessableEntity)
		return
	}
	user.Settings.LocalTimezone, err = time.LoadLocation(form.LocalTimezone)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	err = app.users.UpdateUserSettings(user.ID, user.Settings)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	app.sessionManager.Put(r.Context(), "flash", "Settings changed successfully")
	data := app.newCommonData(r)
	w.Header().Add("HX-Refresh", "true")
	views.UserSettingsView(data, app.timezones).Render(r.Context(), w)

}
