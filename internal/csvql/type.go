package csvql

type Params struct {
	FileInput      string
	DataSourceName string
	Delimiter      string
	Query          string
	Export         string
	Type           string
	Lines          int
}
