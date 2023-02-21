package csvql

import (
	"bytes"
	"database/sql"
	"encoding/csv"
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
	cliPrompt              = "csvql> "
	cliInterruptPrompt     = "^C"
	cliEOFPrompt           = "exit"
	dataSourceNameDefault  = ":memory:"
	sqlCreateTableTemplate = "CREATE TABLE rows (%s\n);"
	sqlInsertTemplate      = "INSERT INTO rows (%s) VALUES (%s);"
)

type Csvql interface {
	Run() error
}

type csvql struct {
	db      *sql.DB
	file    *os.File
	bar     *progressbar.ProgressBar
	params  CsvqlParams
	columns []string
	lines   int
}

func New(params CsvqlParams) Csvql {
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

	if params.DataSourceName == "" {
		params.DataSourceName = dataSourceNameDefault
	}

	return &csvql{params: params, bar: bar}
}

func (c *csvql) Run() error {
	if err := c.loadTotalRows(); err != nil {
		return err
	}

	if err := c.openFile(); err != nil {
		return err
	}
	defer c.file.Close()

	if err := c.openConnection(); err != nil {
		return err
	}
	defer c.db.Close()

	if err := c.loadDataFromFile(); err != nil {
		return err
	}

	if err := c.initializePrompt(); err != nil {
		return err
	}

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
		if err := c.execute(line); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		}
	}

	return nil
}

func (c *csvql) execute(line string) error {
	rows, err := c.db.Query(line)
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

	tbl.Print()

	return nil
}

func (c *csvql) openConnection() error {
	db, err := sql.Open("sqlite3", c.params.DataSourceName)
	if err != nil {
		return err
	}

	c.db = db

	return nil
}

func (c *csvql) openFile() error {
	f, err := os.Open(c.params.FileInput)
	if err != nil {
		return err
	}

	c.file = f

	return nil
}

func (c *csvql) loadTotalRows() error {
	r, err := os.Open(c.params.FileInput)
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

func (c *csvql) loadDataFromFile() error {
	c.bar.ChangeMax(c.lines)
	defer c.bar.Finish()

	r := csv.NewReader(c.file)
	r.Comma = rune(c.params.Delimiter[0])

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

// build table creation statement
func (c *csvql) buildTable() error {
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
func (c *csvql) buildInsert(values []any) error {
	defer c.bar.Add(1)

	columnsRaw := strings.Join(c.columns, ", ")
	paramsRaw := strings.Repeat("?, ", len(c.columns))
	insertRaw := fmt.Sprintf(sqlInsertTemplate, columnsRaw, paramsRaw[:len(paramsRaw)-2])

	if _, err := c.db.Exec(insertRaw, values...); err != nil {
		return err
	}

	return nil
}
