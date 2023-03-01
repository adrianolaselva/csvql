package csv

import (
	"adrianolaselva.github.io/csvql/pkg/importdata"
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"io"
	"os"
	"strings"
)

const (
	sqlCreateTableTemplate = "CREATE TABLE rows (%s\n);"
	sqlInsertTemplate      = "INSERT INTO rows (%s) VALUES (%s);"
)

type csvImport struct {
	db        *sql.DB
	bar       *progressbar.ProgressBar
	file      *os.File
	fileInput string
	columns   []string
	lines     int
	delimiter rune
}

func NewCsvImport(fileInput string, delimiter rune, db *sql.DB, bar *progressbar.ProgressBar) importdata.Import {
	return &csvImport{fileInput: fileInput, delimiter: delimiter, db: db, bar: bar}
}

func (c *csvImport) Import() error {
	c.bar.ChangeMax(c.lines)
	defer c.bar.Finish()

	if err := c.loadTotalRows(); err != nil {
		return err
	}

	return nil
}

func (c *csvImport) loadDataFromFile() error {
	c.bar.ChangeMax(c.lines)
	defer c.bar.Finish()

	r := csv.NewReader(c.file)
	r.Comma = c.delimiter

	headers, err := r.Read()
	if err != nil {
		return err
	}

	c.columns = headers
	if err := c.buildTable(); err != nil {
		return err
	}

	for {
		records, err := r.Read()
		if err == io.EOF {
			break
		}

		var values []any
		for _, r := range records {
			values = append(values, r)
		}

		if err := c.buildInsert(values); err != nil {
			return err
		}
	}

	return nil
}

func (c *csvImport) Close() error {
	return nil
}

func (c *csvImport) openFile() error {
	f, err := os.Open(c.fileInput)
	if err != nil {
		return err
	}

	c.file = f

	return nil
}

func (c *csvImport) loadTotalRows() error {
	r, err := os.Open(c.fileInput)
	if err != nil {
		return err
	}
	defer r.Close()

	buf := make([]byte, 32*1024)
	c.lines = 0
	lineSep := []byte{'\n'}

	for {
		r, err := r.Read(buf)
		c.lines += bytes.Count(buf[:r], lineSep)

		switch {
		case err == io.EOF:
			return nil

		case err != nil:
			return err
		}
	}
}

// build table creation statement
func (c *csvImport) buildTable() error {
	defer c.bar.Add(1)

	var tableAttrsRaw strings.Builder
	for ln, v := range c.columns {
		tableAttrsRaw.WriteString(fmt.Sprintf("\n\t%s text", v))
		if len(c.columns)-1 > ln {
			tableAttrsRaw.WriteString(",")
		}
	}

	if _, err := c.db.Exec(fmt.Sprintf(sqlCreateTableTemplate, tableAttrsRaw.String())); err != nil {
		return err
	}

	return nil
}

// build insert create statement
func (c *csvImport) buildInsert(values []any) error {
	defer c.bar.Add(1)

	columnsRaw := strings.Join(c.columns, ", ")
	paramsRaw := strings.Repeat("?, ", len(c.columns))
	insertRaw := fmt.Sprintf(sqlInsertTemplate, columnsRaw, paramsRaw[:len(paramsRaw)-2])

	if _, err := c.db.Exec(insertRaw, values...); err != nil {
		return err
	}

	return nil
}
