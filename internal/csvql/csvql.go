package csvql

import (
	"adrianolaselva.github.io/csvql/pkg/exportdata/jsonl"
	"adrianolaselva.github.io/csvql/pkg/filehandler"
	csvHandler "adrianolaselva.github.io/csvql/pkg/filehandler/csv"
	"adrianolaselva.github.io/csvql/pkg/storage"
	"adrianolaselva.github.io/csvql/pkg/storage/sqlite"
	"database/sql"
	"fmt"
	"github.com/chzyer/readline"
	"github.com/fatih/color"
	_ "github.com/mattn/go-sqlite3"
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
	params      CsvqlParams
	fileHandler filehandler.FileHandler
}

func New(params CsvqlParams) (Csvql, error) {
	sqLiteStorage, err := sqlite.NewSqLiteStorage(params.DataSourceName)
	if err != nil {
		return nil, err
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

	impData := csvHandler.NewCsvHandler(params.FileInput, rune(params.Delimiter[0]), bar, sqLiteStorage)

	return &csvql{params: params, bar: bar, fileHandler: impData, storage: sqLiteStorage}, nil
}

func (c *csvql) Run() error {
	defer c.bar.Clear()

	if err := c.fileHandler.Import(); err != nil {
		return fmt.Errorf("failed to import data %s", err)
	}
	defer c.fileHandler.Close()

	return c.execute()
}

func (c *csvql) execute() error {
	if c.params.Query != "" && c.params.Export == "" {
		return c.executeQuery(c.params.Query)
	}

	if c.params.Query != "" && c.params.Export != "" {
		return c.executeQueryAndExport(c.params.Query)
	}

	if err := c.initializePrompt(); err != nil {
		return err
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
		return err
	}
	defer l.Close()

	l.CaptureExitSignal()

	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			}

			continue
		}

		if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		if err := c.executeQuery(line); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		}
	}

	return nil
}

func (c *csvql) executeQueryAndExport(line string) error {
	c.bar.Reset()
	c.bar.ChangeMax(c.fileHandler.Lines())
	defer c.bar.Finish()

	rows, err := c.storage.Query(line)
	if err != nil {
		return err
	}
	defer rows.Close()

	return jsonl.NewJsonlExport(rows, c.params.Export, c.bar).Export()
}

func (c *csvql) executeQuery(line string) error {
	rows, err := c.storage.Query(line)
	if err != nil {
		return err
	}
	defer rows.Close()

	return c.printResult(rows)
}

func (c *csvql) printResult(rows *sql.Rows) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
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
			return err
		}

		tbl.AddRow(values...)
	}

	_ = c.bar.Clear()
	tbl.Print()

	return nil
}
