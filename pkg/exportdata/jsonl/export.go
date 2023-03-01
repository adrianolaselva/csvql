package jsonl

import (
	"adrianolaselva.github.io/csvql/pkg/exportdata"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"os"
	"path/filepath"
)

const (
	fileModeDefault os.FileMode = 0644
)

type jsonlExport struct {
	rows       *sql.Rows
	bar        *progressbar.ProgressBar
	exportPath string
}

func NewJsonlExport(rows *sql.Rows, exportPath string, bar *progressbar.ProgressBar) exportdata.Export {
	return &jsonlExport{rows: rows, exportPath: exportPath, bar: bar}
}

func (j *jsonlExport) Export() error {
	columns, err := j.rows.Columns()
	if err != nil {
		return err
	}

	if err = os.MkdirAll(filepath.Dir(j.exportPath), os.ModePerm); err != nil {
		return err
	}

	bufferRx, err := os.OpenFile(j.exportPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, fileModeDefault)
	defer bufferRx.Close()

	for j.rows.Next() {
		j.bar.Add(1)
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}

		if err := j.rows.Scan(pointers...); err != nil {
			return err
		}

		attr := map[string]interface{}{}
		for i, c := range columns {
			attr[c] = pointers[i]
		}

		payload, err := json.Marshal(attr)
		if err != nil {
			return err
		}

		buffer := new(bytes.Buffer)
		if err := json.Compact(buffer, payload); err != nil {
			return err
		}

		if _, err := bufferRx.WriteString(fmt.Sprintf("%s\n", buffer.String())); err != nil {
			return err
		}
	}

	return nil
}
