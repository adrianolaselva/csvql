package storage

import "database/sql"

type Storage interface {
	BuildStructure(string, []string) error
	InsertRow(string, []string, []any) error
	Query(cmd string) (*sql.Rows, error)
	ShowTables() (*sql.Rows, error)
	Close() error
}
