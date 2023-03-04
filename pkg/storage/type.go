package storage

import "database/sql"

type Storage interface {
	SetTableName(string) Storage
	SetColumns([]string) Storage
	BuildStructure() error
	InsertRow([]any) error
	Query(cmd string) (*sql.Rows, error)
	Close() error
}
