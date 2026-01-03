package models

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

type InstallModel struct {
	DB *sql.DB
}

type InstallModelInterface interface {
	SetupDatabase() error
	SetInstalledFlag() error
	GetInstalledFlag() (bool, error)
}

func (m *InstallModel) SetupDatabase() error {
	driver, err := sqlite3.WithInstance(m.DB, &sqlite3.Config{})
	if err != nil {
		return err
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	mFolder := fmt.Sprintf("file://%s/migrations", wd)
	fmt.Printf("migrations %s\n", mFolder)
	mi, err := migrate.NewWithDatabaseInstance("file://migrations", "sqlite", driver)
	if err != nil {
		return err
	}
	err = mi.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func (m *InstallModel) SetInstalledFlag() error {
	stmt := `INSERT INTO installed (InstallDate) VALUES (?)`
	_, err := m.DB.Exec(stmt, time.Now().UTC())
	if err != nil {
		return err
	}
	return nil
}

func (m *InstallModel) GetInstalledFlag() (bool, error) {
	stmt := `SELECT InstallDate FROM installed`
	row := m.DB.QueryRow(stmt)
	var d time.Time
	err := row.Scan(&d)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}
