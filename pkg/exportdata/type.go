package exportdata

type Export interface {
	Export() error
	Close() error
}
