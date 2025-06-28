package transaction

import (
	"database/sql"
	"log/slog"
)

type SqlTransactionManager struct {
	db *sql.DB
}

type TransactionManagerInterface interface {
	RunWithTransaction(callback func(tx *sql.Tx) (interface{}, error)) (interface{}, error)
}

func New(db *sql.DB) *SqlTransactionManager {
	return &SqlTransactionManager{db: db}
}

func (m *SqlTransactionManager) RunWithTransaction(callback func(tx *sql.Tx) (interface{}, error)) (interface{}, error) {
	slog.Info("[*] executing transaction")
	tx, err := m.db.Begin()
	if err != nil {
		return nil, err
	}

	callbackResponse, err := callback(tx)
	if err != nil {
		tx.Rollback()
		slog.Info("[*] rollback executed")
		return callbackResponse, err
	}

	tx.Commit()
	slog.Info("[*] commit executed")
	return callbackResponse, nil
}
