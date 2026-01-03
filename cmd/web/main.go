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
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"git.32bit.cafe/32bitcafe/guestbook/internal/auth"
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
	oidcConfig *oidc.Config
	config     auth.OAuth2ConfigInterface
	provider   *oidc.Provider
	verifier   *oidc.IDTokenVerifier
}

type applicationConfig struct {
	oauthEnabled     bool
	localAuthEnabled bool
	oauth            applicationOauthConfig
	rootUrl          string
	environment      string
}

type application struct {
	sequence          uint16
	logger            *slog.Logger
	websites          models.WebsiteModelInterface
	users             models.UserModelInterface
	guestbookComments models.GuestbookCommentModelInterface
	sessionManager    *scs.SessionManager
	formDecoder       *schema.Decoder
	config            applicationConfig
	debug             bool
	timezones         []string
}

type appInstaller struct {
	app            *application
	srv            *http.Server
	installModel   models.InstallModelInterface
	migrationsPath string
}

func main() {
	addr := flag.String("addr", ":3000", "HTTP network address")
	dsn := flag.String("dsn", "guestbook.db", "data source name")
	migrations := flag.String("migrations", "migrations", "folder containing sql migrations")
	debug := flag.Bool("debug", false, "enable debug mode")
	env := flag.String("env", ".env", ".env file path")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	err := godotenv.Load(*env)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	cfg, err := setupConfig(*addr)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

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
		config:            cfg,
		debug:             *debug,
		timezones:         getAvailableTimezones(),
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	installer := &appInstaller{
		app:            app,
		installModel:   &models.InstallModel{DB: db},
		migrationsPath: filepath.Clean(*migrations),
	}
	installer.srv = &http.Server{
		Addr:         *addr,
		Handler:      installer.installRoutes(),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	err = runInstaller(installer)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

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

func setupConfig(addr string) (applicationConfig, error) {
	var c applicationConfig

	var (
		rootUrl           = os.Getenv("ROOT_URL")
		oidcEnabled       = os.Getenv("ENABLE_OIDC")
		localLoginEnabled = os.Getenv("ENABLE_LOCAL_LOGIN")
		oauth2Provider    = os.Getenv("OAUTH2_PROVIDER")
		clientID          = os.Getenv("OAUTH2_CLIENT_ID")
		clientSecret      = os.Getenv("OAUTH2_CLIENT_SECRET")
		environment       = os.Getenv("ENVIRONMENT")
	)

	if environment != "" {
		c.environment = environment
	} else {
		c.environment = "DEV"
	}

	if rootUrl != "" {
		c.rootUrl = rootUrl
	} else {
		u, err := url.Parse(fmt.Sprintf("https://localhost%s", addr))
		if err != nil {
			return c, err
		}
		c.rootUrl = u.String()
	}

	oauthEnabled, err := strconv.ParseBool(oidcEnabled)
	if err != nil {
		c.oauthEnabled = false
	}
	c.oauthEnabled = oauthEnabled

	localAuthEnabled, err := strconv.ParseBool(localLoginEnabled)
	if err != nil {
		c.localAuthEnabled = true
	}
	c.localAuthEnabled = localAuthEnabled

	if !c.oauthEnabled && !c.localAuthEnabled {
		return c, errors.New("Either ENABLE_OIDC or ENABLE_LOCAL_LOGIN must be set to true")
	}

	// if OIDC is disabled, no more configuration needs to be read
	if !oauthEnabled {
		return c, nil
	}

	var o applicationOauthConfig
	if oauth2Provider == "" || clientID == "" || clientSecret == "" {
		return c, errors.New("OAUTH2_PROVIDER, OAUTH2_CLIENT_ID, and OAUTH2_CLIENT_SECRET must be specified as environment variables.")
	}

	o.ctx = context.Background()
	provider, err := oidc.NewProvider(o.ctx, oauth2Provider)
	if err != nil {
		return c, err
	}
	o.provider = provider
	o.oidcConfig = &oidc.Config{
		ClientID: clientID,
	}
	o.verifier = provider.Verifier(o.oidcConfig)
	o.config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  fmt.Sprintf("%s/users/login/oidc/callback", c.rootUrl),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	c.oauth = o

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

func runInstaller(i *appInstaller) error {
	i.app.logger.Info("Performing migrations")
	err := i.installModel.SetupDatabase(i.migrationsPath)
	if err != nil {
		return err
	}
	installed, _ := i.installModel.GetInstalledFlag()
	if installed {
		return nil
	}
	i.app.logger.Info("App not installed, running installer...")
	i.app.logger.Info("Starting installation server", slog.Any("addr", i.srv.Addr))
	if i.app.debug {
		err = i.srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	} else {
		err = i.srv.ListenAndServe()
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	i.app.logger.Info("Installation complete")
	err = i.installModel.SetInstalledFlag()
	if err != nil {
		return err
	}
	return nil
}
