package memory

import (
	"github.com/itimofeev/go-saga/storage"
	"github.com/juju/errors"
)

func New() storage.Storage {
	s, err := newMemStorage()
	if err != nil {
		panic(err)
	}
	return s
}

type memStorage struct {
	data map[string][]*storage.Log
}

// NewMemStorage creates log storage base on memory.
// This storage use simple `map[string][]string`, just for TestCase used.
// NOT use this in product.
func newMemStorage() (storage.Storage, error) {
	return &memStorage{
		data: make(map[string][]*storage.Log),
	}, nil
}

// AppendLog appends log into queue under given logID.
func (s *memStorage) AppendLog(log *storage.Log) error {
	logQueue, ok := s.data[log.SagaLogID]
	if !ok {
		logQueue = []*storage.Log{}
		s.data[log.SagaLogID] = logQueue
	}
	s.data[log.SagaLogID] = append(s.data[log.SagaLogID], log)
	return nil
}

// Lookup lookups log under given logID.
func (s *memStorage) Lookup(logID string) ([]*storage.Log, error) {
	return s.data[logID], nil
}

// Close uses to close storage and release resources.
func (s *memStorage) Close() error {
	return nil
}

func (s *memStorage) Cleanup(logID string) error {
	delete(s.data, logID)
	return nil
}

func (s *memStorage) LastLog(logID string) (*storage.Log, error) {
	logData, ok := s.data[logID]
	if !ok {
		err := errors.NewErr("LogData %s not found", logID)
		return nil, &err
	}
	sizeOfLog := len(logData)
	if sizeOfLog == 0 {
		return nil, errors.New("LogData is empty")
	}
	lastLog := logData[sizeOfLog-1]
	return lastLog, nil
}
