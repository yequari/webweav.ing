package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"

	"git.32bit.cafe/32bitcafe/guestbook/internal/models"
	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/schema"
	_ "github.com/mattn/go-sqlite3"
)

type application struct {
	sequence          uint16
	logger            *slog.Logger
	websites          *models.WebsiteModel
	guestbooks        *models.GuestbookModel
	users             *models.UserModel
	guestbookComments *models.GuestbookCommentModel
	sessionManager    *scs.SessionManager
	formDecoder       *schema.Decoder
	debug             bool
	timezones         []string
}

func main() {
	addr := flag.String("addr", ":3000", "HTTP network address")
	dsn := flag.String("dsn", "guestbook.db", "data source name")
	debug := flag.Bool("debug", false, "enable debug mode")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	db, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	sessionManager := scs.New()
	sessionManager.Store = sqlite3store.New(db)
	sessionManager.Lifetime = 12 * time.Hour

	formDecoder := schema.NewDecoder()
	formDecoder.IgnoreUnknownKeys(true)

	app := &application{
		sequence:          0,
		logger:            logger,
		sessionManager:    sessionManager,
		websites:          &models.WebsiteModel{DB: db},
		guestbooks:        &models.GuestbookModel{DB: db},
		users:             &models.UserModel{DB: db, Settings: make(map[string]models.Setting)},
		guestbookComments: &models.GuestbookCommentModel{DB: db},
		formDecoder:       formDecoder,
		debug:             *debug,
		timezones:         getAvailableTimezones(),
	}

	err = app.users.InitializeSettingsMap()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	err = app.guestbooks.InitializeSettingsMap()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:         *addr,
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Info("Starting server", slog.Any("addr", *addr))

	if app.debug {
		err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	} else {
		err = srv.ListenAndServe()
	}
	logger.Error(err.Error())
	os.Exit(1)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func getAvailableTimezones() []string {
	var zones []string
	var zoneDirs = []string{
		"/usr/share/zoneinfo/",
		"/usr/share/lib/zoneinfo/",
		"/usr/lib/locale/TZ/",
	}
	for _, zd := range zoneDirs {
		zones = walkTzDir(zd, zones)
		for idx, zone := range zones {
			zones[idx] = strings.ReplaceAll(zone, zd+"/", "")
		}
	}
	return zones
}

func walkTzDir(path string, zones []string) []string {
	fileInfos, err := os.ReadDir(path)
	if err != nil {
		return zones
	}
	isAlpha := func(s string) bool {
		for _, r := range s {
			if !unicode.IsLetter(r) {
				return false
			}
		}
		return true
	}
	for _, info := range fileInfos {
		if info.Name() != strings.ToUpper(info.Name()[:1])+info.Name()[1:] {
			continue
		}
		if !isAlpha(info.Name()[:1]) {
			continue
		}
		newPath := path + "/" + info.Name()
		if info.IsDir() {
			zones = walkTzDir(newPath, zones)
		} else {
			zones = append(zones, newPath)
		}
	}
	return zones

}
