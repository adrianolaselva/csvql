package filehandler

type FileHandler interface {
	Import() error
	Lines() int
	Close() error
}
