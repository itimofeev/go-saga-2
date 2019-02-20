package postgres

import (
	"fmt"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/itimofeev/go-saga/storage"
	"time"
)

func New() storage.Storage {
	s, err := newRmdbStorage()
	if err != nil {
		panic(err)
	}
	return s
}

type rmdbStorage struct {
	db *pg.DB
}

type TXLog struct {
	LogID      string `sql:",pk,notnull"`
	SubTxID    string `sql:",pk,notnull"`
	Data       []storage.ParamData
	CreateTime time.Time
	Type       storage.LogType `sql:",pk,notnull"`
}

type hook struct {
}

func (*hook) BeforeQuery(event *pg.QueryEvent) {
	query, err := event.FormattedQuery()
	if err != nil {
		panic(err)
	}
	fmt.Println(query)
}

func (*hook) AfterQuery(*pg.QueryEvent) {
}

// newRmdbStorage creates log storage base on rmdb.
func newRmdbStorage() (storage.Storage, error) {
	opts, err := pg.ParseURL("postgresql://postgres:@db:5432/postgres?sslmode=disable")
	if err != nil {
		return nil, err
	}
	db := pg.Connect(opts)

	db.AddQueryHook(&hook{})

	// Migrate the schema
	if err := createSchema(db); err != nil {
		return nil, err
	}

	//_, _ = db.Exec(`ALTER TABLE offers ADD CONSTRAINT offer__user_task__unique UNIQUE (user_id, task_id)`)

	return &rmdbStorage{db: db}, nil
}

func createSchema(db *pg.DB) error {
	for _, mdl := range []interface{}{
		(*TXLog)(nil),
	} {
		err := db.CreateTable(mdl, &orm.CreateTableOptions{
			IfNotExists:   true,
			FKConstraints: true,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// AppendLog appends log into queue under given logID.
func (s *rmdbStorage) AppendLog(log *storage.Log) error {
	txlog := &TXLog{
		LogID:      log.SagaLogID,
		SubTxID:    log.SubTxID,
		Data:       log.Params,
		CreateTime: log.Time,
		Type:       log.Type,
	}
	return s.db.Insert(txlog)
}

// Lookup lookups log under given logID.
func (s *rmdbStorage) Lookup(logID string) (datas []*storage.Log, err error) {
	var txLogs []*TXLog
	err = s.db.Model(&txLogs).
		Where("log_id = ?", logID).
		Select()

	for _, txLog := range txLogs {
		datas = append(datas, &storage.Log{
			SagaLogID: txLog.LogID,
			SubTxID:   txLog.SubTxID,
			Time:      txLog.CreateTime,
			Type:      txLog.Type,
			Params:    txLog.Data,
		})
	}

	return datas, err
}

// Close uses to close storage and release resources.
func (s *rmdbStorage) Close() error {
	return s.db.Close()
}

func (s *rmdbStorage) Cleanup(logID string) error {
	_, err := s.db.Model((*TXLog)(nil)).Where("log_id = ?", logID).Delete()
	return err
}

func (s *rmdbStorage) LastLog(logID string) (data *storage.Log, err error) {
	var txLog TXLog
	err = s.db.Model(&txLog).
		Where("log_id = ?", logID).
		Order("create_time DESC").
		Select()

	if err != nil {
		return nil, err
	}

	return &storage.Log{
		SagaLogID: txLog.LogID,
		SubTxID:   txLog.SubTxID,
		Time:      txLog.CreateTime,
		Type:      txLog.Type,
		Params:    txLog.Data,
	}, nil
}
