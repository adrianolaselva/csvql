package exportdata

import (
	"adrianolaselva.github.io/csvql/pkg/exportdata"
	"adrianolaselva.github.io/csvql/pkg/exportdata/csv"
	"adrianolaselva.github.io/csvql/pkg/exportdata/jsonl"
	"database/sql"
	"fmt"
	"github.com/schollz/progressbar/v3"
)

const (
	CSVLineExportType  = "csv"
	JSONLineExportType = "jsonl"
)

func NewExport(exportType string, rows *sql.Rows, exportPath string, bar *progressbar.ProgressBar) (exportdata.Export, error) {
	switch exportType {
	case CSVLineExportType:
		return csv.NewCsvExport(rows, exportPath, bar), nil
	case JSONLineExportType:
		return jsonl.NewJsonlExport(rows, exportPath, bar), nil
	}

	return nil, fmt.Errorf("export type %s not defined", exportType)
}
