package storage

// Storage uses to support save and lookup saga log.
type Storage interface {

	// AppendLog appends log data into log under given logID
	AppendLog(log *Log) error

	// Lookup uses to lookup all log under given logID
	Lookup(logID string) ([]*Log, error)

	// Close use to close storage and release resources
	Close() error

	// Cleanup cleans up all log data in logID
	Cleanup(logID string) error

	// LastLog fetch last log entry with given logID
	LastLog(logID string) (*Log, error)
}
