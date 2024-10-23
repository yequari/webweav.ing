package main

import (
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"text/template"
	"time"

	"git.32bit.cafe/32bitcafe/guestbook/internal/models"
	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type application struct {
    logger *slog.Logger
    templateCache map[string]*template.Template
    guestbooks *models.GuestbookModel
    users *models.UserModel
    guestbookComments *models.GuestbookCommentModel
    sessionManager *scs.SessionManager
}

func main() {
    addr := flag.String("addr", ":3000", "HTTP network address")
    dsn := flag.String("dsn", "guestbook.db", "data source name")
    flag.Parse()

    logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

    db, err := openDB(*dsn)
    if err != nil {
        logger.Error(err.Error())
        os.Exit(1)
    }
    defer db.Close()

    templateCache, err := newTemplateCache()
    if err != nil {
        logger.Error(err.Error())
        os.Exit(1)
    }

    sessionManager := scs.New()
    sessionManager.Store = sqlite3store.New(db)
    sessionManager.Lifetime = 12 * time.Hour

    app := &application{
        templateCache: templateCache,
        logger: logger,
        sessionManager: sessionManager,
        guestbooks: &models.GuestbookModel{DB: db},
        users: &models.UserModel{DB: db},
        guestbookComments: &models.GuestbookCommentModel{DB: db},
    }

    logger.Info("Starting server", slog.Any("addr", *addr))

    err = http.ListenAndServe(*addr, app.routes());
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

func getUserId() uuid.UUID {
    userId, _ := decodeIdB64("laINnbnkTtyN5SYoCfSbXw")
    return userId
}
