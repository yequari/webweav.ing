package main

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"git.32bit.cafe/32bitcafe/guestbook/internal/models"
	"git.32bit.cafe/32bitcafe/guestbook/ui/views"
	"github.com/gorilla/schema"
	"github.com/justinas/nosurf"
)

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	app.logger.Error(err.Error(), "method", method, "uri", uri)
	if app.debug {
		http.Error(w, err.Error()+"\n"+string(debug.Stack()), http.StatusInternalServerError)
	}
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) nextSequence() uint16 {
	val := app.sequence
	if app.sequence == math.MaxUint16 {
		app.sequence = 0
	} else {
		app.sequence += 1
	}
	return val
}

func (app *application) createShortId() uint64 {
	now := time.Now().UTC()
	epoch, err := time.Parse(time.RFC822Z, "01 Jan 20 00:00 -0000")
	if err != nil {
		fmt.Println(err)
		return 0
	}
	d := now.Sub(epoch)
	ms := d.Milliseconds()
	seq := app.nextSequence()
	return (uint64(ms) & 0x0FFFFFFFFFFFFFFF) | (uint64(seq) << 48)
}

func shortIdToSlug(id uint64) string {
	slug := strconv.FormatUint(id, 36)
	return slug
}

func slugToShortId(slug string) uint64 {
	id, _ := strconv.ParseUint(slug, 36, 64)
	return id
}

func (app *application) decodePostForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		var multiErrors *schema.MultiError
		if !errors.As(err, &multiErrors) {
			panic(err)
		}
		return err
	}
	return nil
}

func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}
	return isAuthenticated
}

func (app *application) getCurrentUser(r *http.Request) *models.User {
	if !app.isAuthenticated(r) {
		return nil
	}
	user, ok := r.Context().Value(userNameContextKey).(models.User)
	if !ok {
		return nil
	}
	return &user
}

func (app *application) newCommonData(r *http.Request) views.CommonData {
	return views.CommonData{
		CurrentYear:     time.Now().Year(),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		IsAuthenticated: app.isAuthenticated(r),
		CSRFToken:       nosurf.Token(r),
		CurrentUser:     app.getCurrentUser(r),
		IsHtmx:          r.Header.Get("Hx-Request") == "true",
	}
}

func DefaultUserSettings() models.UserSettings {
	return models.UserSettings{
		LocalTimezone: time.Now().UTC().Location(),
	}
}

func (app *application) durationToTime(duration string) (time.Time, error) {
	var result time.Time
	offset, err := time.ParseDuration(duration)
	if err != nil {
		return result, nil
	}
	result = time.Now().UTC().Add(offset)
	return result, nil
}
