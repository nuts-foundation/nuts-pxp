package db

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/glebarez/sqlite"
	"github.com/nuts-foundation/nuts-pxp/config"
	sql_migrations "github.com/nuts-foundation/nuts-pxp/db/migrations"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

var _ DB = (*SqlDB)(nil)

type SqlDB struct {
	sqlDB *gorm.DB
}

func New(config config.Config) (*SqlDB, error) {
	e := &SqlDB{}
	connectionString := config.SQL.ConnectionString
	if len(connectionString) == 0 {
		return nil, errors.New("no SQL connection string provided")
	}

	// Find right SQL adapter for ORM and migrations
	dbType := strings.Split(connectionString, ":")[0]
	if dbType == "sqlite" {
		connectionString = connectionString[strings.Index(connectionString, ":")+1:]
	} else if dbType == "mysql" {
		// MySQL DSN needs to be without mysql://
		// See https://github.com/go-sql-driver/mysql#examples
		connectionString = strings.TrimPrefix(connectionString, "mysql://")
	}
	db, err := goose.OpenDBWithDriver(dbType, connectionString)
	if err != nil {
		return nil, err
	}
	var dialect goose.Dialect
	gormConfig := &gorm.Config{
		TranslateError: true,
	}
	// SQL migration files use env variables for substitutions.
	// TEXT SQL data type is really DB-specific, so we set a default here and override it for a specific database type (MS SQL).
	_ = os.Setenv("TEXT_TYPE", "TEXT")
	defer os.Unsetenv("TEXT_TYPE")
	switch dbType {
	case "sqlite":
		// SQLite does not support SELECT FOR UPDATE and allows only 1 active write transaction at any time,
		// and any other attempt to acquire a write transaction will directly return an error.
		// This is in contrast to most other SQL-databases, which let the 2nd thread wait for some time to acquire the lock.
		// The general advice for SQLite is to retry the operation, which is just poor-man's scheduling.
		// So to keep behavior consistent across databases, we'll just limit the number connections to 1 if it's a SQLite store.
		// With 1 connection, all actions will be performed sequentially. This impacts performance, but SQLite should not be used in production.
		// See https://github.com/nuts-foundation/nuts-node/pull/2589#discussion_r1399130608
		db.SetMaxOpenConns(1)
		dialector := sqlite.Dialector{Conn: db}
		e.sqlDB, err = gorm.Open(dialector, gormConfig)
		if err != nil {
			return nil, err
		}
		dialect = goose.DialectSQLite3
	case "mysql":
		e.sqlDB, _ = gorm.Open(mysql.New(mysql.Config{
			Conn: db,
		}), gormConfig)
		dialect = goose.DialectMySQL
	case "postgres":
		e.sqlDB, _ = gorm.Open(postgres.New(postgres.Config{
			Conn: db,
		}), gormConfig)
		dialect = goose.DialectPostgres
	case "sqlserver":
		_ = os.Setenv("TEXT_TYPE", "VARCHAR(MAX)")
		e.sqlDB, _ = gorm.Open(sqlserver.New(sqlserver.Config{
			Conn: db,
		}), gormConfig)
		dialect = goose.DialectMSSQL
	default:
		return nil, errors.New("unsupported SQL database")
	}

	gooseProvider, err := goose.NewProvider(dialect, db, sql_migrations.SQLMigrationsFS)
	if err != nil {
		return nil, err
	}

	_, err = gooseProvider.Up(context.Background())
	if err != nil && !errors.Is(err, goose.ErrNoNextVersion) {
		return nil, err
	}

	return e, nil
}

func (db *SqlDB) Close() error {
	sqlDB, err := db.sqlDB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (db *SqlDB) Create(data SQLData) error {
	return db.sqlDB.Create(data).Error
}

func (db *SqlDB) Delete(id string) error {
	return db.sqlDB.Where("id = ?", id).Delete(&SQLData{}).Error
}

func (db *SqlDB) Get(id string) (SQLData, error) {
	var record SQLData
	err := db.sqlDB.Model(&SQLData{}).Where("id = ?", id).First(&record).Error
	if err != nil {
		return SQLData{}, err
	}
	return record, nil
}

func (db *SqlDB) Query(scope string, verifier string, client string) (string, error) {
	var record SQLData
	// todo multiple records
	err := db.sqlDB.Model(&SQLData{}).Where("scope = ? AND verifier = ? AND client = ?", scope, verifier, client).
		First(&record).Error
	if err != nil {
		return "", err
	}
	return record.AuthInput, nil
}
