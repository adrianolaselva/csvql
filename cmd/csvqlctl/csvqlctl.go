package csvqlctl

import (
	"adrianolaselva.github.io/csvql/internal/csvql"
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
)

type CsvQlCtl interface {
	Command() (*cobra.Command, error)
	runE(cmd *cobra.Command, args []string) error
}

type csvQlCtl struct {
	rootCmd *cobra.Command
	params  csvql.CsvqlParams
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
		StringVarP(&c.params.FileInput, fileParam, fileShortParam, "", "origin file in csv")

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

	if err := command.MarkPersistentFlagRequired(fileParam); err != nil {
		return nil, err
	}

	return command, nil
}

func (c *csvQlCtl) runE(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true
	csvql, err := csvql.New(c.params)
	if err != nil {
		return err
	}
	defer csvql.Close()

	return csvql.Run()
}
