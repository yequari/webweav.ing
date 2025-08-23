package main

import (
	"net/http"

	"git.32bit.cafe/32bitcafe/guestbook/internal/forms"
	"git.32bit.cafe/32bitcafe/guestbook/internal/models"
	"git.32bit.cafe/32bitcafe/guestbook/internal/validator"
	"git.32bit.cafe/32bitcafe/guestbook/ui"
	"git.32bit.cafe/32bitcafe/guestbook/ui/views"
	"github.com/justinas/alice"
)

func (i *appInstaller) installRoutes() http.Handler {
	mux := http.NewServeMux()
	if i.app.config.environment == "PROD" {
		mux.Handle("GET /static/", http.FileServerFS(ui.Files))
	} else {
		fileServer := http.FileServer(http.Dir("./ui/static/"))
		mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))
	}

	mux.HandleFunc("GET /ping", ping)
	standard := alice.New(i.app.recoverPanic, i.app.logRequest, commonHeaders)
	mux.Handle("/{$}", standard.ThenFunc(i.getInstallHomepage))
	mux.Handle("GET /install", standard.ThenFunc(i.getInstallForm))
	mux.Handle("POST /install", standard.ThenFunc(i.postInstallForm))

	return standard.Then(mux)
}

func (i *appInstaller) getInstallHomepage(w http.ResponseWriter, r *http.Request) {
	views.InitialInstallView("Installation").Render(r.Context(), w)
}

func (i *appInstaller) getInstallForm(w http.ResponseWriter, r *http.Request) {
	var form forms.InstallForm
	views.InstallFormView("Installation - Settings", form).Render(r.Context(), w)
}

func (i *appInstaller) postInstallForm(w http.ResponseWriter, r *http.Request) {
	var form forms.InstallForm
	err := i.app.decodePostForm(r, &form)
	if err != nil {
		i.app.clientError(w, http.StatusBadRequest)
		return
	}
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")
	if !form.Valid() {
		w.WriteHeader(http.StatusUnprocessableEntity)
		views.InstallFormView("User Registration", form).Render(r.Context(), w)
		return
	}
	sId := i.app.createShortId()
	err = i.app.users.Insert(sId, form.Name, form.Email, form.Password, DefaultUserSettings())
	if err != nil {
		i.app.serverError(w, r, err)
	}
	err = i.app.users.AddUserToGroup(1, models.AdminGroup)
	if err != nil {
		i.app.serverError(w, r, err)
	}
	views.InstallSuccessView().Render(r.Context(), w)
	i.srv.Close()
}
