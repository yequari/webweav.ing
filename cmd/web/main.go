package main

import (
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"text/template"

	"git.32bit.cafe/32bitcafe/guestbook/internal/models"
	_ "modernc.org/sqlite"
)

type application struct {
    logger *slog.Logger
    templateCache map[string]*template.Template
    guestbooks *models.GuestbookModel
    users *models.UserModel
    guestbookComments *models.GuestbookCommentModel
}

func main() {
    addr := flag.String("addr", ":3000", "HTTP network address")
    dsn := flag.String("dsn", "guestbook.db", "data source name")
    flag.Parse()

    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

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

    app := &application{
        templateCache: templateCache,
        logger: logger,
        guestbooks: &models.GuestbookModel{DB: db},
        users: &models.UserModel{DB: db},
        guestbookComments: &models.GuestbookCommentModel{DB: db},
    }

    logger.Info("Starting server on %s", slog.Any("addr", ":4000"))

    err = http.ListenAndServe(*addr, app.routes());
    logger.Error(err.Error())
    os.Exit(1)
}

func openDB(dsn string) (*sql.DB, error) {
    db, err := sql.Open("sqlite", dsn)
    if err != nil {
        return nil, err
    }
    if err = db.Ping(); err != nil {
        return nil, err
    }
    return db, nil
}
