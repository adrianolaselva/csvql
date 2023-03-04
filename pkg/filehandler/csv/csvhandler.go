package csv

import (
	"adrianolaselva.github.io/csvql/pkg/filehandler"
	"adrianolaselva.github.io/csvql/pkg/storage"
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"io"
	"os"
)

type csvHandler struct {
	bar       *progressbar.ProgressBar
	storage   storage.Storage
	file      *os.File
	fileInput string
	lines     int
	delimiter rune
}

func NewCsvHandler(fileInput string, delimiter rune, bar *progressbar.ProgressBar, storage storage.Storage) filehandler.FileHandler {
	return &csvHandler{fileInput: fileInput, delimiter: delimiter, storage: storage, bar: bar}
}

// Import import data
func (c *csvHandler) Import() error {
	if err := c.openFile(); err != nil {
		return err
	}

	if err := c.loadTotalRows(); err != nil {
		return err
	}

	if err := c.loadDataFromFile(); err != nil {
		return err
	}

	return nil
}

// Query execute statements
func (c *csvHandler) Query(cmd string) (*sql.Rows, error) {
	return c.storage.Query(cmd)
}

// Lines return total lines
func (c *csvHandler) Lines() int {
	return c.lines
}

// Close execute in defer
func (c *csvHandler) Close() error {
	defer func(storage storage.Storage) {
		_ = storage.Close()
	}(c.storage)

	defer func(file *os.File) {
		_ = file.Close()
	}(c.file)

	return nil
}

func (c *csvHandler) loadDataFromFile() error {
	c.bar.ChangeMax(c.lines)

	r := csv.NewReader(c.file)
	r.Comma = c.delimiter

	headers, err := r.Read()
	if err != nil {
		return fmt.Errorf("failed to load headers: %s", err)
	}

	if err := c.storage.SetColumns(headers).BuildStructure(); err != nil {
		return fmt.Errorf("failed to load headers and build structure: %s", err)
	}

	line := 1
	for {
		line++
		records, err := r.Read()
		if err == io.EOF {
			break
		}

		var values []any
		for _, r := range records {
			values = append(values, r)
		}

		_ = c.bar.Add(1)
		if err := c.storage.InsertRow(values); err != nil {
			return fmt.Errorf("failed to process row: %s", err)
		}
	}

	return nil
}

// openFile open file
func (c *csvHandler) openFile() error {
	f, err := os.Open(c.fileInput)
	if err != nil {
		return fmt.Errorf("failed to open file: %s", err)
	}

	c.file = f

	return nil
}

// loadTotalRows load total rows in file
func (c *csvHandler) loadTotalRows() error {
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
			return fmt.Errorf("failed to totalize rows: %s", err)
		}
	}
}
