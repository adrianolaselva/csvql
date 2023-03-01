package importdata

type Import interface {
	Import() error
	Close() error
}
