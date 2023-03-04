package sqlite

import (
	"adrianolaselva.github.io/csvql/pkg/storage"
	"database/sql"
	"fmt"
	"strings"
)

const (
	sqlCreateTableTemplate = "CREATE TABLE rows (%s\n);"
	sqlInsertTemplate      = "INSERT INTO %s (%s) VALUES (%s);"
	defaultTableName       = "rows"
	dataSourceNameDefault  = ":memory:"
)

type sqLiteStorage struct {
	db        *sql.DB
	tableName string
	columns   []string
}

func NewSqLiteStorage(datasource string) (storage.Storage, error) {
	if datasource == "" {
		datasource = dataSourceNameDefault
	}

	db, err := sql.Open("sqlite3", datasource)
	if err != nil {
		return nil, err
	}

	return &sqLiteStorage{db: db, tableName: defaultTableName}, nil
}

func (s *sqLiteStorage) SetTableName(tableName string) storage.Storage {
	s.tableName = tableName
	return s
}

func (s *sqLiteStorage) SetColumns(columns []string) storage.Storage {
	s.columns = columns
	return s
}

// BuildStructure build table creation statement
func (s *sqLiteStorage) BuildStructure() error {
	var tableAttrsRaw strings.Builder
	for i, v := range s.columns {
		tableAttrsRaw.WriteString(fmt.Sprintf("\n\t%s text", v))
		if len(s.columns)-1 > i {
			tableAttrsRaw.WriteString(",")
		}
	}

	query := fmt.Sprintf(sqlCreateTableTemplate, tableAttrsRaw.String())
	if _, err := s.db.Exec(query); err != nil {
		return fmt.Errorf("failed to create structure: %s (sql: %s)", err, query)
	}

	return nil
}

// InsertRow build insert create statement
func (s *sqLiteStorage) InsertRow(values []any) error {
	columnsRaw := strings.Join(s.columns, ", ")
	paramsRaw := strings.Repeat("?, ", len(s.columns))
	query := fmt.Sprintf(sqlInsertTemplate, s.tableName, columnsRaw, paramsRaw[:len(paramsRaw)-2])

	if _, err := s.db.Exec(query, values...); err != nil {
		return fmt.Errorf("failed to execute insert: %s (sql: %s)", err, query)
	}

	return nil
}

// Query execute statements
func (s *sqLiteStorage) Query(cmd string) (*sql.Rows, error) {
	return s.db.Query(cmd)
}

// Close execute in defer
func (s *sqLiteStorage) Close() error {
	return s.db.Close()
}
