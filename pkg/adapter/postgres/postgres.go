package databaseconnection

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"os"

	_ "github.com/lib/pq"
)

type ApplicationDatabaseInterface interface {
	Client() *sql.DB
	DataBaseHealth() error
}

type ApplicationDatabase struct {
	sqlClient *sql.DB
}

func NewDbConnection() *ApplicationDatabase {
	connStr := fmt.Sprintf(
		"%s://%s:%s@%s/%s%s",
		os.Getenv("DB_DRIVER"),
		os.Getenv("DB_USER"),
		url.QueryEscape(os.Getenv("DB_PASS")),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_OPTIONS"),
	)

	client, err := sql.Open(os.Getenv("DB_DRIVER"), connStr)
	if err != nil {
		slog.Error("erro ao conectar com banco de dados", err.Error(), err)
		panic(err)
	}

	if err = client.Ping(); err != nil {
		slog.Error("erro testar banco de dados", err.Error(), err)
		panic(err)
	}

	slog.Info("banco de dados conectado com sucesso!")

	return &ApplicationDatabase{sqlClient: client}
}

func (s *ApplicationDatabase) Client() *sql.DB {
	return s.sqlClient
}

func (s *ApplicationDatabase) DataBaseHealth() error {
	return s.sqlClient.Ping()
}
