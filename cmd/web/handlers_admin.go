package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"git.32bit.cafe/32bitcafe/guestbook/internal/forms"
	"git.32bit.cafe/32bitcafe/guestbook/internal/models"
	"git.32bit.cafe/32bitcafe/guestbook/internal/validator"
	"git.32bit.cafe/32bitcafe/guestbook/ui/views"
)

func (app *application) getAdminPanelLanding(w http.ResponseWriter, r *http.Request) {
	websites, err := app.websites.GetCount()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	users, err := app.users.GetCount()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	comments, err := app.guestbookComments.GetCount()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	stats := views.AdminStat{
		WebsiteCount: websites,
		UserCount:    users,
		CommentCount: comments,
	}
	data := app.newCommonData(r)
	views.AdminPanelLandingView("Admin Panel", data, stats).Render(r.Context(), w)
}

func (app *application) getAdminPanelAllUsers(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")
	count := r.URL.Query().Get("count")
	var pageNum int64 = 1
	var pageSize int64 = 5
	var err error
	if page != "" {
		pageNum, err = strconv.ParseInt(page, 10, 0)
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}
	if count != "" {
		pageSize, err = strconv.ParseInt(count, 10, 0)
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}
	users, err := app.users.GetAllPage(pageNum, pageSize)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	total, err := app.users.GetCount()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newCommonData(r)
	views.AdminPanelUsersView("All Users - Admin", data, users, pageNum, pageSize, total).Render(r.Context(), w)
}

func (app *application) getAdminPanelUser(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	u, err := app.users.Get(slugToShortId(slug))
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	data := app.newCommonData(r)
	views.AdminPanelUserMgmtView(fmt.Sprintf("User Management - %s", u.Username), data, u).Render(r.Context(), w)
}

func (app *application) getAdminPanelUserMgmtDetail(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	u, err := app.users.Get(slugToShortId(slug))
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	commonData := app.newCommonData(r)
	views.AdminPanelUserMgmtDetail(commonData.CSRFToken, u).Render(r.Context(), w)
}

func (app *application) getAdminPanelUserMgmtForm(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	u, err := app.users.Get(slugToShortId(slug))
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	var form forms.AdminUserMgmtForm
	form.Username = u.Username
	form.Email = u.Email

	data := app.newCommonData(r)
	views.AdminPanelUserMgmtEditForm(data.CSRFToken, form, u, []models.UserGroupId{}).Render(r.Context(), w)
}

func (app *application) putAdminPanelUserMgmtForm(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	u, err := app.users.Get(slugToShortId(slug))
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	var form forms.AdminUserMgmtForm
	err = app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
	}

	form.CheckField(validator.NotBlank(form.Username), "admin_username", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "admin_useremail", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "admin_useremail", "Please provide a valid email address")
	if !form.Valid() {
		data := app.newCommonData(r)
		w.WriteHeader(http.StatusUnprocessableEntity)
		views.AdminPanelUserMgmtEditForm(data.CSRFToken, form, u, []models.UserGroupId{}).Render(r.Context(), w)
		return
	}
	updatedUser := u
	updatedUser.Username = form.Username
	updatedUser.Email = form.Email
	err = app.users.UpdateUser(updatedUser)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	commonData := app.newCommonData(r)
	views.AdminPanelUserMgmtDetail(commonData.CSRFToken, updatedUser).Render(r.Context(), w)

}

func (app *application) putAdminPanelBanUser(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	u, err := app.users.Get(slugToShortId(slug))
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	err = app.users.BanUser(u.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	u.Banned = time.Now()
	commonData := app.newCommonData(r)
	views.AdminPanelUserMgmtDetail(commonData.CSRFToken, u).Render(r.Context(), w)
}

func (app *application) putAdminPanelUnbanUser(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	u, err := app.users.Get(slugToShortId(slug))
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	err = app.users.UnbanUser(u.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	u.Banned = time.Time{}
	commonData := app.newCommonData(r)
	views.AdminPanelUserMgmtDetail(commonData.CSRFToken, u).Render(r.Context(), w)
}

func (app *application) getAdminPanelWebsites(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")
	count := r.URL.Query().Get("count")
	var pageNum int64 = 1
	var pageSize int64 = 5
	var err error
	if page != "" {
		pageNum, err = strconv.ParseInt(page, 10, 0)
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}
	if count != "" {
		pageSize, err = strconv.ParseInt(count, 10, 0)
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}
	websites, err := app.websites.GetAllPage(pageNum, pageSize)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	total, err := app.websites.GetCount()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	commonData := app.newCommonData(r)
	views.AdminPanelAllWebsitesView("All websites", commonData, websites, pageNum, pageSize, total).Render(r.Context(), w)
}

func (app *application) getAdminPanelWebsiteDetails(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	page := r.URL.Query().Get("page")
	count := r.URL.Query().Get("count")
	var pageNum int64 = 1
	var pageSize int64 = 25
	var err error
	if page != "" {
		pageNum, err = strconv.ParseInt(page, 10, 0)
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}
	if count != "" {
		pageSize, err = strconv.ParseInt(count, 10, 0)
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
	}
	website, err := app.websites.Get(slugToShortId(slug))
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	total, err := app.guestbookComments.GetAllCount(website.Guestbook.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	comments, err := app.guestbookComments.GetAllPage(website.Guestbook.ID, pageNum, pageSize)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	commonData := app.newCommonData(r)
	views.AdminPanelWebsiteDetailView(fmt.Sprintf("Admin - %s", website.Name), commonData, website, comments, pageNum, pageSize, total).Render(r.Context(), w)
}
