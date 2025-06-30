package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"

	"git.32bit.cafe/32bitcafe/guestbook/internal/models"
	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/schema"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"
)

type applicationOauthConfig struct {
	ctx        context.Context
	config     oauth2.Config
	provider   *oidc.Provider
	oidcConfig *oidc.Config
	verifier   *oidc.IDTokenVerifier
}

type application struct {
	sequence          uint16
	logger            *slog.Logger
	websites          models.WebsiteModelInterface
	users             models.UserModelInterface
	guestbookComments models.GuestbookCommentModelInterface
	sessionManager    *scs.SessionManager
	formDecoder       *schema.Decoder
	oauth             applicationOauthConfig
	debug             bool
	timezones         []string
	rootUrl           string
}

func main() {
	addr := flag.String("addr", ":3000", "HTTP network address")
	dsn := flag.String("dsn", "guestbook.db", "data source name")
	debug := flag.Bool("debug", false, "enable debug mode")
	root := flag.String("root", "https://localhost:3000", "root URL of application")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	godotenv.Load(".env.dev")

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
		users:             &models.UserModel{DB: db, Settings: make(map[string]models.Setting)},
		guestbookComments: &models.GuestbookCommentModel{DB: db},
		formDecoder:       formDecoder,
		debug:             *debug,
		timezones:         getAvailableTimezones(),
		rootUrl:           *root,
	}

	o, err := setupOauth(app.rootUrl)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	app.oauth = o

	err = app.users.InitializeSettingsMap()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	err = app.websites.InitializeSettingsMap()
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

func setupOauth(rootUrl string) (applicationOauthConfig, error) {
	var c applicationOauthConfig
	var (
		oauth2Provider = os.Getenv("OAUTH2_PROVIDER")
		clientID       = os.Getenv("OAUTH2_CLIENT_ID")
		clientSecret   = os.Getenv("OAUTH2_CLIENT_SECRET")
	)
	if oauth2Provider == "" || clientID == "" || clientSecret == "" {
		return applicationOauthConfig{}, errors.New("OAUTH2_PROVIDER, OAUTH2_CLIENT_ID, and OAUTH2_CLIENT_SECRET must be specified as environment variables.")
	}

	c.ctx = context.Background()
	provider, err := oidc.NewProvider(c.ctx, oauth2Provider)
	if err != nil {
		return applicationOauthConfig{}, err
	}
	c.provider = provider
	c.oidcConfig = &oidc.Config{
		ClientID: clientID,
	}
	c.verifier = provider.Verifier(c.oidcConfig)
	c.config = oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  fmt.Sprintf("%s/users/login/oidc/callback", rootUrl),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	return c, nil

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
