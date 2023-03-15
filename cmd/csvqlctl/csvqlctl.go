package csvqlctl

import (
	"adrianolaselva.github.io/csvql/internal/csvql"
	"fmt"
	"github.com/spf13/cobra"
)

const (
	fileParam               = "file"
	fileShortParam          = "f"
	fileDelimiterParam      = "delimiter"
	fileShortDelimiterParam = "d"
	queryParam              = "query"
	queryShortParam         = "q"
	storageParam            = "storage"
	storageShortParam       = "s"
	exportParam             = "export"
	exportShortParam        = "e"
	typeParam               = "type"
	typeShortParam          = "t"
	linesParam              = "lines"
	linesShortParam         = "l"
	tableNameParam          = "collection"
	tableNameShortParam     = "c"
)

type CsvQlCtl interface {
	Command() (*cobra.Command, error)
	runE(cmd *cobra.Command, args []string) error
}

type csvQlCtl struct {
	params csvql.Params
}

func New() CsvQlCtl {
	return &csvQlCtl{}
}

func (c *csvQlCtl) Command() (*cobra.Command, error) {
	command := &cobra.Command{
		Use:     "run",
		Short:   "Load and run queries from csv file",
		Long:    `./csvql run -f test.csv -d ";"`,
		Example: `./csvql run -f test.csv -d ";"`,
		RunE:    c.runE,
	}

	command.
		PersistentFlags().
		StringArrayVarP(&c.params.FileInputs, fileParam, fileShortParam, []string{}, "origin file in csv")

	command.
		PersistentFlags().
		StringVarP(&c.params.Delimiter, fileDelimiterParam, fileShortDelimiterParam, ",", "csv delimiter")

	command.
		PersistentFlags().
		StringVarP(&c.params.Query, queryParam, queryShortParam, "", "query param")

	command.
		PersistentFlags().
		StringVarP(&c.params.Export, exportParam, exportShortParam, "", "export path")

	command.
		PersistentFlags().
		StringVarP(&c.params.Type, typeParam, typeShortParam, "", "format type [`jsonl`,`csv`]")

	command.
		PersistentFlags().
		StringVarP(&c.params.DataSourceName, storageParam, storageShortParam, "", "sqlite file")

	command.
		PersistentFlags().
		IntVarP(&c.params.Lines, linesParam, linesShortParam, 0, "number of lines to be read")

	if err := command.MarkPersistentFlagRequired(fileParam); err != nil {
		return nil, fmt.Errorf("failed to validate flag %s: %w", fileParam, err)
	}

	if c.params.Export != "" && c.params.Type == "" {
		return nil, fmt.Errorf("failed to validate flag")
	}

	return command, nil
}

func (c *csvQlCtl) runE(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true
	csvQl, err := csvql.New(c.params)
	if err != nil {
		return fmt.Errorf("failed to initialize csvql: %w", err)
	}
	defer func(csvQl csvql.Csvql) {
		_ = csvQl.Close()
	}(csvQl)

	err = csvQl.Run()
	if err != nil {
		return fmt.Errorf("failed to run csvql: %w", err)
	}

	return nil
}
