package csvql

import (
	"adrianolaselva.github.io/csvql/internal/exportdata"
	"adrianolaselva.github.io/csvql/pkg/filehandler"
	csvHandler "adrianolaselva.github.io/csvql/pkg/filehandler/csv"
	"adrianolaselva.github.io/csvql/pkg/storage"
	"adrianolaselva.github.io/csvql/pkg/storage/sqlite"
	"database/sql"
	"errors"
	"fmt"
	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/schollz/progressbar/v3"
	"io"
	"os"
	"strings"
)

const (
	cliPrompt          = "csvql> "
	cliInterruptPrompt = "^C"
	cliEOFPrompt       = "exit"
)

type Csvql interface {
	Run() error
	Close() error
}

type csvql struct {
	storage     storage.Storage
	bar         *progressbar.ProgressBar
	params      Params
	fileHandler filehandler.FileHandler
}

func New(params Params) (Csvql, error) {
	sqLiteStorage, err := sqlite.NewSqLiteStorage(params.DataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	bar := progressbar.NewOptions(0,
		progressbar.OptionSetWriter(os.Stdout),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetDescription("[cyan][1/1][reset] loading data..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	impData := csvHandler.NewCsvHandler(params.FileInputs, rune(params.Delimiter[0]), bar, sqLiteStorage, params.Lines)

	return &csvql{params: params, bar: bar, fileHandler: impData, storage: sqLiteStorage}, nil
}

// Run import file content and run command
func (c *csvql) Run() error {
	defer func(bar *progressbar.ProgressBar) {
		_ = bar.Clear()
	}(c.bar)

	if err := c.fileHandler.Import(); err != nil {
		return fmt.Errorf("failed to import data %w", err)
	}
	defer func(fileHandler filehandler.FileHandler) {
		_ = fileHandler.Close()
	}(c.fileHandler)

	rows, err := c.storage.ShowTables()
	if err != nil {
		return fmt.Errorf("failed to list tables: %w", err)
	}

	if err := c.printResult(rows); err != nil {
		return fmt.Errorf("failed to print tables: %w", err)
	}

	return c.execute()
}

// execute execution after data import
func (c *csvql) execute() error {
	switch {
	case c.params.Query != "" && c.params.Export == "":
		return c.executeQuery(c.params.Query)
	case c.params.Query != "" && c.params.Export != "":
		return c.executeQueryAndExport(c.params.Query)
	default:
		if err := c.initializePrompt(); err != nil {
			return err
		}
	}

	return nil
}

func (c *csvql) Close() error {
	defer func(fileHandler filehandler.FileHandler) {
		_ = fileHandler.Close()
	}(c.fileHandler)

	return nil
}

func (c *csvql) initializePrompt() error {
	l, err := readline.NewEx(&readline.Config{
		Prompt:          cliPrompt,
		InterruptPrompt: cliInterruptPrompt,
		EOFPrompt:       cliEOFPrompt,
		AutoComplete: readline.SegmentFunc(func(i [][]rune, i2 int) [][]rune {
			return nil
		}),
	})
	if err != nil {
		return fmt.Errorf("failed to initialize cli: %w", err)
	}

	defer func(l *readline.Instance) {
		_ = l.Close()
	}(l)

	l.CaptureExitSignal()

	for {
		line, err := l.Readline()
		if errors.Is(err, readline.ErrInterrupt) {
			if len(line) == 0 {
				break
			}

			continue
		}

		if errors.Is(err, io.EOF) {
			break
		}

		line = strings.TrimSpace(line)
		if err := c.executeQuery(line); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		}
	}

	return nil
}

// executeQueryAndExport execute query and export
func (c *csvql) executeQueryAndExport(line string) error {
	c.bar.Reset()
	c.bar.ChangeMax(c.fileHandler.Lines())
	defer func(bar *progressbar.ProgressBar) {
		_ = bar.Finish()
	}(c.bar)

	rows, err := c.storage.Query(line)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	export, err := exportdata.NewExport(c.params.Type, rows, c.params.Export, c.bar)
	if err != nil {
		return fmt.Errorf("failed to export: %w", err)
	}

	if err := export.Export(); err != nil {
		return fmt.Errorf("failed to export data: %w", err)
	}

	_ = c.bar.Clear()

	fmt.Printf("[%s] file successfully exported\n", c.params.Export)

	return nil
}

func (c *csvql) executeQuery(line string) error {
	rows, err := c.storage.Query(line)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	return c.printResult(rows)
}

func (c *csvql) printResult(rows *sql.Rows) error {
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to load columns: %w", err)
	}

	cols := make([]interface{}, 0)
	for _, c := range columns {
		cols = append(cols, c)
	}

	tbl := table.New(cols...).
		WithHeaderFormatter(color.New(color.FgGreen, color.Underline).SprintfFunc()).
		WithFirstColumnFormatter(color.New(color.FgYellow).SprintfFunc()).
		WithWriter(os.Stdout)

	for rows.Next() {
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}

		if err := rows.Scan(pointers...); err != nil {
			return fmt.Errorf("failed to read row: %w", err)
		}

		tbl.AddRow(values...)
	}

	_ = c.bar.Clear()
	tbl.Print()

	return nil
}
