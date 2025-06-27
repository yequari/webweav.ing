package main

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"git.32bit.cafe/32bitcafe/guestbook/internal/assert"
)

func TestPing(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, body := ts.get(t, "/ping")

	assert.Equal(t, code, http.StatusOK)
	assert.Equal(t, body, "OK")
}

func TestGetGuestbookView(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody string
	}{
		{
			name:     "Valid id",
			urlPath:  fmt.Sprintf("/websites/%s/guestbook", shortIdToSlug(1)),
			wantCode: http.StatusOK,
			wantBody: "Guestbook for Example",
		},
		{
			name:     "Non-existent ID",
			urlPath:  fmt.Sprintf("/websites/%s/guestbook", shortIdToSlug(2)),
			wantCode: http.StatusNotFound,
		},
		{
			name:     "String ID",
			urlPath:  "/websites/abcd/guestbook",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)
			assert.Equal(t, code, tt.wantCode)
			if tt.wantBody != "" {
				assert.StringContains(t, body, tt.wantBody)
			}
		})
	}
}

func TestPostGuestbookCommentCreate(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	_, _, body := ts.get(t, fmt.Sprintf("/websites/%s/guestbook", shortIdToSlug(1)))
	validCSRFToken := extractCSRFToken(t, body)

	const (
		validAuthorName  = "John Test"
		validAuthorEmail = "test@example.com"
		validAuthorSite  = "example.com"
		validContent     = "This is a comment"
	)

	tests := []struct {
		name        string
		authorName  string
		authorEmail string
		authorSite  string
		content     string
		csrfToken   string
		wantCode    int
	}{
		{
			name:        "Valid input",
			authorName:  validAuthorName,
			authorEmail: validAuthorEmail,
			authorSite:  validAuthorSite,
			content:     validContent,
			csrfToken:   validCSRFToken,
			wantCode:    http.StatusSeeOther,
		},
		{
			name:        "Blank name",
			authorName:  "",
			authorEmail: validAuthorEmail,
			authorSite:  validAuthorSite,
			content:     validContent,
			csrfToken:   validCSRFToken,
			wantCode:    http.StatusUnprocessableEntity,
		},
		{
			name:        "Blank email",
			authorName:  validAuthorName,
			authorEmail: "",
			authorSite:  validAuthorSite,
			content:     validContent,
			csrfToken:   validCSRFToken,
			wantCode:    http.StatusSeeOther,
		},
		{
			name:        "Blank site",
			authorName:  validAuthorName,
			authorEmail: validAuthorEmail,
			authorSite:  "",
			content:     validContent,
			csrfToken:   validCSRFToken,
			wantCode:    http.StatusSeeOther,
		},
		{
			name:        "Blank content",
			authorName:  validAuthorName,
			authorEmail: validAuthorEmail,
			authorSite:  validAuthorSite,
			content:     "",
			csrfToken:   validCSRFToken,
			wantCode:    http.StatusUnprocessableEntity,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("authorname", tt.authorName)
			form.Add("authoremail", tt.authorEmail)
			form.Add("authorsite", tt.authorSite)
			form.Add("content", tt.content)
			form.Add("csrf_token", tt.csrfToken)
			code, _, body := ts.postForm(t, fmt.Sprintf("/websites/%s/guestbook/comments/create", shortIdToSlug(1)), form)
			assert.Equal(t, code, tt.wantCode)
			assert.Equal(t, body, body)
		})
	}
}

func TestPostGuestbookCommentCreateRemote(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	_, _, body := ts.get(t, fmt.Sprintf("/websites/%s/guestbook", shortIdToSlug(1)))
	validCSRFToken := extractCSRFToken(t, body)

	const (
		validAuthorName  = "John Test"
		validAuthorEmail = "test@example.com"
		validAuthorSite  = "example.com"
		validContent     = "This is a comment"
	)

	tests := []struct {
		name        string
		authorName  string
		authorEmail string
		authorSite  string
		content     string
		csrfToken   string
		wantCode    int
	}{
		{
			name:        "Valid input",
			authorName:  validAuthorName,
			authorEmail: validAuthorEmail,
			authorSite:  validAuthorSite,
			content:     validContent,
			csrfToken:   validCSRFToken,
			wantCode:    http.StatusSeeOther,
		},
		{
			name:        "Blank name",
			authorName:  "",
			authorEmail: validAuthorEmail,
			authorSite:  validAuthorSite,
			content:     validContent,
			csrfToken:   validCSRFToken,
			wantCode:    http.StatusUnprocessableEntity,
		},
		{
			name:        "Blank email",
			authorName:  validAuthorName,
			authorEmail: "",
			authorSite:  validAuthorSite,
			content:     validContent,
			csrfToken:   validCSRFToken,
			wantCode:    http.StatusSeeOther,
		},
		{
			name:        "Blank site",
			authorName:  validAuthorName,
			authorEmail: validAuthorEmail,
			authorSite:  "",
			content:     validContent,
			csrfToken:   validCSRFToken,
			wantCode:    http.StatusSeeOther,
		},
		{
			name:        "Blank content",
			authorName:  validAuthorName,
			authorEmail: validAuthorEmail,
			authorSite:  validAuthorSite,
			content:     "",
			csrfToken:   validCSRFToken,
			wantCode:    http.StatusUnprocessableEntity,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("authorname", tt.authorName)
			form.Add("authoremail", tt.authorEmail)
			form.Add("authorsite", tt.authorSite)
			form.Add("content", tt.content)
			form.Add("csrf_token", tt.csrfToken)
			code, _, body := ts.postForm(t, fmt.Sprintf("/websites/%s/guestbook/comments/create/remote", shortIdToSlug(1)), form)
			assert.Equal(t, code, tt.wantCode)
			assert.Equal(t, body, body)
		})
	}
}
