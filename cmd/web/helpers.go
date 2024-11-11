package main

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/schema"
)

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
    var (
        method = r.Method
        uri = r.URL.RequestURI()
    )

    app.logger.Error(err.Error(), "method", method, "uri", uri)
    http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
    http.Error(w, http.StatusText(status), status)
}

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
    ts, ok := app.templateCache[page]
    if !ok {
        err := fmt.Errorf("the template %s does not exist", page)
        app.serverError(w, r, err)
        return
    }

    w.WriteHeader(status)
    err := ts.ExecuteTemplate(w, "base", data)
    if err != nil {
        app.serverError(w, r, err)
    }
}

func (app *application) nextSequence () uint16 {
    val := app.sequence
    if app.sequence == math.MaxUint16 {
        app.sequence = 0
    } else {
        app.sequence += 1
    }
    return val
}

func (app *application) createShortId () uint64 {
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
