package main

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/google/uuid"
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

func encodeIdB64 (id uuid.UUID) (string, error) {
    b, err := id.MarshalBinary()
    if err != nil {
        return "", err
    }
    s := base64.RawURLEncoding.EncodeToString(b)
    return s, nil
}

func decodeIdB64 (id string) (uuid.UUID, error) {
    b, err := base64.RawURLEncoding.DecodeString(id)
    var u uuid.UUID
    if err != nil {
        return u, err
    }
    err = u.UnmarshalBinary(b)
    if err != nil {
        return u, err
    }
    return u, nil
}
