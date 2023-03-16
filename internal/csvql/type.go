package csvql

type Params struct {
	FileInputs     []string
	DataSourceName string
	Delimiter      string
	Query          string
	Export         string
	Type           string
	Lines          int
}
