package main

import (
	"context"
	"crypto/rsa"
	"net/http"
	"net/url"
	"testing"
	"time"

	"git.32bit.cafe/32bitcafe/guestbook/internal/assert"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/coreos/go-oidc/v3/oidc/oidctest"
	"golang.org/x/oauth2"
)

func TestUserSignup(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	_, _, body := ts.get(t, "/users/register")
	validCSRFToken := extractCSRFToken(t, body)

	const (
		validName     = "John"
		validPassword = "validPassword"
		validEmail    = "john@example.com"
		formTag       = `<form action="/users/register" method="post">`
	)

	tests := []struct {
		name         string
		userName     string
		userEmail    string
		userPassword string
		csrfToken    string
		wantCode     int
		wantFormTag  string
	}{
		{
			name:         "Valid submission",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusSeeOther,
		},
		{
			name:         "Missing token",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: validPassword,
			wantCode:     http.StatusBadRequest,
		},
		{
			name:         "Empty name",
			userName:     "",
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Empty email",
			userName:     validName,
			userEmail:    "",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Empty password",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: "",
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Invalid email",
			userName:     validName,
			userEmail:    "asdfasdf",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Invalid password",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: "asdfasd",
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Duplicate email",
			userName:     validName,
			userEmail:    "dupe@example.com",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			form := url.Values{}
			form.Add("username", tt.userName)
			form.Add("email", tt.userEmail)
			form.Add("password", tt.userPassword)
			form.Add("csrf_token", tt.csrfToken)
			code, _, body := ts.postForm(t, "/users/register", form)
			assert.Equal(t, code, tt.wantCode)
			if tt.wantFormTag != "" {
				assert.StringContains(t, body, tt.wantFormTag)
			}
		})
	}

}

type OAuth2Mock struct {
	Srv  *testServer
	Priv *rsa.PrivateKey
}

func (o *OAuth2Mock) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return ""
}

func (o *OAuth2Mock) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	tkn := oauth2.Token{
		AccessToken: "AccessToken",
		Expiry:      time.Now().Add(1 * time.Hour),
	}
	m := make(map[string]interface{})
	var rawClaims = `{
		"iss": "` + o.Srv.URL + `",
		"aud": "my-client-id",
		"sub": "foo",
		"email": "foo@example.com",
		"email_verified": true,
		"nonce": "nonce"
		}`

	m["id_token"] = oidctest.SignIDToken(o.Priv, "test-key", oidc.RS256, rawClaims)
	return tkn.WithExtra(m), nil
}

func (o *OAuth2Mock) Client(ctx context.Context, t *oauth2.Token) *http.Client {
	return nil
}

func TestUserOIDCCallback(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	priv := newTestKey(t)
	srv := newTestOIDCServer(t, priv)

	defer srv.Close()
	defer ts.Close()
	ctx := context.Background()

	p, err := oidc.NewProvider(ctx, srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &oidc.Config{
		ClientID:        "my-client-id",
		SkipExpiryCheck: true,
	}
	v := p.VerifierContext(ctx, cfg)
	app.config.oauth = applicationOauthConfig{
		ctx:        context.Background(),
		oidcConfig: cfg,
		config: &OAuth2Mock{
			Srv:  srv,
			Priv: priv,
		},
		provider: p,
		verifier: v,
	}
	app.config.oauthEnabled = true

	const (
		validSubject = "goodSubject"
		validUserId  = 1
		validEmail   = "test@example.com"
		validState   = "goodState"
	)

	tests := []struct {
		name     string
		subject  string
		email    string
		state    string
		wantCode int
	}{
		{
			name:     "Found Subject",
			subject:  validSubject,
			email:    validEmail,
			state:    validState,
			wantCode: http.StatusSeeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			r, err := http.NewRequest("GET", ts.URL, nil)
			if err != nil {
				t.Fatal(err)
			}
			r.URL.Path = "/users/login/oidc/callback"
			q := r.URL.Query()
			q.Add("state", tt.state)
			r.URL.RawQuery = q.Encode()
			c := &http.Cookie{
				Name:     "state",
				Value:    validState,
				MaxAge:   int(time.Hour.Seconds()),
				Secure:   r.TLS != nil,
				HttpOnly: true,
			}
			d := &http.Cookie{
				Name:     "nonce",
				Value:    "nonce",
				MaxAge:   int(time.Hour.Seconds()),
				Secure:   r.TLS != nil,
				HttpOnly: true,
			}
			r.AddCookie(c)
			r.AddCookie(d)
			resp, err := ts.Client().Do(r)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, resp.StatusCode, tt.wantCode)
		})
	}

}
