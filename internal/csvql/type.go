package csvql

type CsvqlParams struct {
	FileInput      string
	DataSourceName string
	Delimiter      string
	Query          string
}

type CsvqlImport interface {
	Import() error
	Close() error
}

type CsvqlExport interface {
	Export() error
}
