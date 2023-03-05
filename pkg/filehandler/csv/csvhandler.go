package csv

import (
	"adrianolaselva.github.io/csvql/pkg/filehandler"
	"adrianolaselva.github.io/csvql/pkg/storage"
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"io"
	"os"
)

const (
	bufferMaxLength = 32 * 1024
)

type csvHandler struct {
	bar         *progressbar.ProgressBar
	storage     storage.Storage
	file        *os.File
	fileInput   string
	totalLines  int
	limitLines  int
	currentLine int
	delimiter   rune
}

func NewCsvHandler(fileInput string, delimiter rune, bar *progressbar.ProgressBar, storage storage.Storage, limitLines int) filehandler.FileHandler {
	return &csvHandler{fileInput: fileInput, delimiter: delimiter, storage: storage, bar: bar, limitLines: limitLines}
}

// Import import data
func (c *csvHandler) Import() error {
	if err := c.openFile(); err != nil {
		return err
	}

	if err := c.loadTotalRows(); err != nil {
		return err
	}

	if c.limitLines > 0 && c.totalLines > c.limitLines {
		c.totalLines = c.limitLines
	}

	if err := c.loadDataFromFile(); err != nil {
		return err
	}

	return nil
}

// Query execute statements
func (c *csvHandler) Query(cmd string) (*sql.Rows, error) {
	rows, err := c.storage.Query(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return rows, nil
}

// Lines return total lines
func (c *csvHandler) Lines() int {
	return c.totalLines
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

// loadDataFromFile load data from file
func (c *csvHandler) loadDataFromFile() error {
	c.bar.ChangeMax(c.totalLines)

	r := csv.NewReader(c.file)
	r.Comma = c.delimiter

	if err := c.readHeader(r); err != nil {
		return fmt.Errorf("failed to load headers and build structure: %w", err)
	}

	c.currentLine = 0
	for {
		err := c.readline(r)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// readHeader read header
func (c *csvHandler) readHeader(r *csv.Reader) error {
	headers, err := r.Read()
	if err != nil {
		return fmt.Errorf("failed to load headers: %w", err)
	}

	if err := c.storage.SetColumns(headers).BuildStructure(); err != nil {
		return fmt.Errorf("failed to load headers and build structure: %w", err)
	}

	return nil
}

// readline read line
func (c *csvHandler) readline(r *csv.Reader) error {
	records, err := r.Read()
	if err != nil {
		return fmt.Errorf("failed to read line: %w", err)
	}

	if c.totalLines == c.currentLine {
		return io.EOF
	}

	_ = c.bar.Add(1)
	c.currentLine++

	if err := c.storage.InsertRow(c.convertToAnyArray(records)); err != nil {
		return fmt.Errorf("failed to process row number %d: %w", c.currentLine, err)
	}

	return nil
}

// convertToAnyArray convert string array to any array
func (c *csvHandler) convertToAnyArray(records []string) []any {
	values := make([]any, 0, len(records))
	for _, r := range records {
		values = append(values, r)
	}

	return values
}

// openFile open file
func (c *csvHandler) openFile() error {
	f, err := os.Open(c.fileInput)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	c.file = f

	return nil
}

// loadTotalRows load total rows in file
func (c *csvHandler) loadTotalRows() error {
	r, err := os.Open(c.fileInput)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", c.fileInput, err)
	}
	defer r.Close()

	buf := make([]byte, bufferMaxLength)
	c.totalLines = 0
	lineSep := []byte{'\n'}

	for {
		r, err := r.Read(buf)
		c.totalLines += bytes.Count(buf[:r], lineSep)

		switch {
		case err == io.EOF:
			return nil

		case err != nil:
			return fmt.Errorf("failed to totalize rows: %w", err)
		}
	}
}
