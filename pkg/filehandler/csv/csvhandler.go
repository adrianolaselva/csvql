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
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

const (
	bufferMaxLength = 32 * 1024
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

type csvHandler struct {
	mx          sync.Mutex
	bar         *progressbar.ProgressBar
	storage     storage.Storage
	files       []*os.File
	fileInputs  []string
	totalLines  int
	limitLines  int
	currentLine int
	delimiter   rune
}

func NewCsvHandler(fileInputs []string, delimiter rune, bar *progressbar.ProgressBar, storage storage.Storage, limitLines int) filehandler.FileHandler {
	return &csvHandler{fileInputs: fileInputs, delimiter: delimiter, storage: storage, bar: bar, limitLines: limitLines}
}

// Import import data
func (c *csvHandler) Import() error {
	if err := c.openFiles(); err != nil {
		return err
	}

	wg := new(sync.WaitGroup)
	wg.Add(len(c.fileInputs))
	errChannels := make(chan error, len(c.fileInputs))

	for _, file := range c.fileInputs {
		go func(wg *sync.WaitGroup, file string, errChan chan error) {
			defer wg.Done()
			err := c.loadTotalRows(file)
			errChan <- err
		}(wg, file, errChannels)
	}

	wg.Wait()
	if err := <-errChannels; err != nil {
		return err
	}

	if c.limitLines > 0 && c.totalLines > c.limitLines {
		c.totalLines = c.limitLines
	}

	wg.Add(len(c.files))
	errChannels = make(chan error, len(c.files))
	for _, file := range c.files {
		tableName := c.formatTableName(file)
		go func(wg *sync.WaitGroup, file *os.File, tableName string, errChan chan error) {
			defer wg.Done()
			errChan <- c.loadDataFromFile(tableName, file)
		}(wg, file, tableName, errChannels)
	}

	wg.Wait()
	if err := <-errChannels; err != nil {
		return err
	}

	return nil
}

// formatTableName format table name by removing invalid characters
func (c *csvHandler) formatTableName(file *os.File) string {
	tableName := strings.ReplaceAll(strings.ToLower(filepath.Base(file.Name())), filepath.Ext(file.Name()), "")
	tableName = strings.ReplaceAll(tableName, " ", "_")
	return nonAlphanumericRegex.ReplaceAllString(tableName, "")
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

	defer func(files []*os.File) {
		for _, file := range files {
			_ = file.Close()
		}
	}(c.files)

	return nil
}

// loadDataFromFile load data from file
func (c *csvHandler) loadDataFromFile(tableName string, file *os.File) error {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.bar.ChangeMax(c.totalLines)

	r := csv.NewReader(file)
	r.Comma = c.delimiter

	columns, err := c.readHeader(tableName, r)
	if err != nil {
		return fmt.Errorf("failed to load headers and build structure: %w", err)
	}

	c.currentLine = 0
	for {
		err := c.readline(tableName, columns, r)
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
func (c *csvHandler) readHeader(tableName string, r *csv.Reader) ([]string, error) {
	columns, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to load headers: %w", err)
	}

	if err := c.storage.BuildStructure(tableName, columns); err != nil {
		return nil, fmt.Errorf("failed to load headers and build structure: %w", err)
	}

	return columns, nil
}

// readline read line
func (c *csvHandler) readline(tableName string, columns []string, r *csv.Reader) error {
	records, err := r.Read()
	if err != nil {
		return fmt.Errorf("failed to read line: %w", err)
	}

	if c.totalLines == c.currentLine {
		return io.EOF
	}

	_ = c.bar.Add(1)
	c.currentLine++

	if err := c.storage.InsertRow(tableName, columns, c.convertToAnyArray(records)); err != nil {
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
func (c *csvHandler) openFiles() error {
	wg := new(sync.WaitGroup)
	wg.Add(len(c.fileInputs))
	errChannels := make(chan error, len(c.fileInputs))

	for _, file := range c.fileInputs {
		go func(wg *sync.WaitGroup, file string, errChan chan error) {
			defer wg.Done()

			f, err := os.Open(file)
			if err != nil {
				errChan <- fmt.Errorf("failed to open file: %w", err)
				return
			}

			c.files = append(c.files, f)
			errChan <- nil
		}(wg, file, errChannels)
	}

	wg.Wait()
	if err := <-errChannels; err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	return nil
}

// loadTotalRows load total rows in file
func (c *csvHandler) loadTotalRows(file string) error {
	r, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", file, err)
	}
	defer func(r *os.File) {
		_ = r.Close()
	}(r)

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
