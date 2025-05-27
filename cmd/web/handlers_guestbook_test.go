package main

import (
	"fmt"
	"net/http"
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
