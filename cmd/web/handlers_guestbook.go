package main

import (
	"encoding/json"
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

func (app *application) getGuestbook(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	website, err := app.websites.Get(slugToShortId(slug))
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	if !website.Guestbook.Settings.IsVisible {
		u := app.getCurrentUser(r)
		if u == nil {
			app.clientError(w, http.StatusForbidden)
			return
		}
		if u.ID != website.UserId {
			app.clientError(w, http.StatusForbidden)
			return
		}
	}
	comments, err := app.guestbookComments.GetVisible(website.Guestbook.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newCommonData(r)
	views.GuestbookView(website.Name, data, website, website.Guestbook, comments, forms.CommentCreateForm{}).Render(r.Context(), w)
}

func (app *application) getGuestbookComments(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	website, err := app.websites.Get(slugToShortId(slug))
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	comments, err := app.guestbookComments.GetAll(website.Guestbook.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := app.newCommonData(r)
	views.GuestbookDashboardCommentsView(fmt.Sprintf("%s - Comments", website.Name), data, website, website.Guestbook, comments).Render(r.Context(), w)
}

func (app *application) getGuestbookCommentsSerialized(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	website, err := app.websites.Get(slugToShortId(slug))
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	if !website.Guestbook.Settings.IsVisible || !website.Guestbook.Settings.AllowRemoteHostAccess {
		app.clientError(w, http.StatusForbidden)
	}
	comments, err := app.guestbookComments.GetVisibleSerialized(website.Guestbook.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	b, err := json.Marshal(comments)
	w.Write(b)
}

func (app *application) getGuestbookCommentCreate(w http.ResponseWriter, r *http.Request) {
	// TODO: This will be the embeddable form
	slug := r.PathValue("id")
	website, err := app.websites.Get(slugToShortId(slug))
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	s := website.Guestbook.Settings
	if !s.IsVisible || !s.AllowRemoteHostAccess || !website.Guestbook.CanComment() {
		app.clientError(w, http.StatusForbidden)
	}
	data := app.newCommonData(r)
	form := forms.CommentCreateForm{}
	views.EmbeddableGuestbookCommentForm(data, website, form).Render(r.Context(), w)
}

func (app *application) postGuestbookCommentCreate(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	headless, err := strconv.ParseBool(r.URL.Query().Get("headless"))
	if err != nil {
		headless = false
	}
	website, err := app.websites.Get(slugToShortId(slug))
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	if !website.Guestbook.CanComment() {
		app.clientError(w, http.StatusForbidden)
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
	form.CheckField(validator.MaxChars(form.AuthorEmail, 256), "authorEmail", "This field cannot be more than 256 characters long")
	form.CheckField(validator.MaxChars(form.AuthorSite, 256), "authorSite", "This field cannot be more than 256 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")

	if !form.Valid() {
		data := app.newCommonData(r)
		w.WriteHeader(http.StatusUnprocessableEntity)
		if headless {
			views.EmbeddableGuestbookCommentForm(data, website, form).Render(r.Context(), w)
		}
		// TODO: use htmx to avoid getting comments again
		comments, err := app.guestbookComments.GetVisible(website.Guestbook.ID)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		views.GuestbookView("Guestbook", data, website, website.Guestbook, comments, form).Render(r.Context(), w)
		return
	}

	shortId := app.createShortId()
	_, err = app.guestbookComments.Insert(shortId, website.Guestbook.ID, 0, form.AuthorName, form.AuthorEmail, form.AuthorSite, form.Content, "", true)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	app.sessionManager.Put(r.Context(), "flash", "Comment successfully posted!")
	if headless {
		http.Redirect(w, r, fmt.Sprintf("/websites/%s/guestbook/comments/create", slug), http.StatusSeeOther)
	}
	http.Redirect(w, r, fmt.Sprintf("/websites/%s/guestbook", slug), http.StatusSeeOther)
}

func (app *application) postGuestbookCommentCreateRemote(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	website, err := app.websites.Get(slugToShortId(slug))
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	if !matchOrigin(r.Header.Get("Origin"), website.Url) {
		app.clientError(w, http.StatusForbidden)
		return
	}
	if !website.Guestbook.CanComment() {
		app.clientError(w, http.StatusForbidden)
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
	form.CheckField(validator.MaxChars(form.AuthorEmail, 256), "authorEmail", "This field cannot be more than 256 characters long")
	form.CheckField(validator.MaxChars(form.AuthorSite, 256), "authorSite", "This field cannot be more than 256 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")

	// if redirect path is filled out, redirect to that path on the website host
	// otherwise redirect to the guestbook by default
	redirectUrl := fmt.Sprintf("/websites/%s/guestbook", shortIdToSlug(website.ShortId))
	if form.Redirect != "" {
		u, err := website.Url.Parse(form.Redirect)
		if err == nil {
			redirectUrl = u.String()
		}
	}

	if !form.Valid() {
		views.GuestbookCommentCreateRemoteErrorView(redirectUrl, "Invalid Input").Render(r.Context(), w)
		return
	}

	shortId := app.createShortId()
	_, err = app.guestbookComments.Insert(shortId, website.Guestbook.ID, 0, form.AuthorName, form.AuthorEmail, form.AuthorSite, form.Content, "", true)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	views.GuestbookCommentCreateRemoteSuccessView(redirectUrl).Render(r.Context(), w)
}

func (app *application) getCommentQueue(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	website, err := app.websites.Get(slugToShortId(slug))
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	comments, err := app.guestbookComments.GetAll(website.Guestbook.ID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data := app.newCommonData(r)
	views.GuestbookDashboardCommentsView(fmt.Sprintf("%s - Comment Queue", website.Name), data, website, website.Guestbook, comments).Render(r.Context(), w)
}

func (app *application) getCommentTrash(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	website, err := app.websites.Get(slugToShortId(slug))
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	comments, err := app.guestbookComments.GetDeleted(website.Guestbook.ID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data := app.newCommonData(r)
	views.GuestbookDashboardCommentsView(fmt.Sprintf("%s - Comment Trash", website.Name), data, website, website.Guestbook, comments).Render(r.Context(), w)
}

func (app *application) putHideGuestbookComment(w http.ResponseWriter, r *http.Request) {
	user := app.getCurrentUser(r)
	wSlug := r.PathValue("id")
	website, err := app.websites.Get(slugToShortId(wSlug))
	if err != nil {
		app.logger.Info("website 404")
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	if user.ID != website.UserId {
		app.clientError(w, http.StatusUnauthorized)
	}
	cSlug := r.PathValue("commentId")
	comment, err := app.guestbookComments.Get(slugToShortId(cSlug))
	if err != nil {
		app.logger.Info("comment 404")
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	comment.IsPublished = !comment.IsPublished
	err = app.guestbookComments.UpdateComment(&comment)
	if err != nil {
		app.serverError(w, r, err)
	}
	data := app.newCommonData(r)
	views.GuestbookDashboardUpdateButtonPart(data, website, comment).Render(r.Context(), w)
}

func (app *application) deleteGuestbookComment(w http.ResponseWriter, r *http.Request) {
	user := app.getCurrentUser(r)
	wSlug := r.PathValue("id")
	website, err := app.websites.Get(slugToShortId(wSlug))
	if err != nil {
		app.logger.Info("website 404")
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	if user.ID != website.UserId {
		app.clientError(w, http.StatusUnauthorized)
	}
	cSlug := r.PathValue("commentId")
	comment, err := app.guestbookComments.Get(slugToShortId(cSlug))
	if err != nil {
		app.logger.Info("comment 404")
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	comment.Deleted = time.Now().UTC()
	err = app.guestbookComments.UpdateComment(&comment)
	if err != nil {
		app.serverError(w, r, err)
	}
	views.GuestbookDashboardCommentDeletePart("Comment was successfully deleted").Render(r.Context(), w)
}

func (app *application) getAllGuestbooks(w http.ResponseWriter, r *http.Request) {
	websites, err := app.websites.GetAll()
	if err != nil {
		app.serverError(w, r, err)
	}
	views.AllGuestbooksView(app.newCommonData(r), websites).Render(r.Context(), w)
}
