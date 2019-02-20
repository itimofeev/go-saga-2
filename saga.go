// Package saga provide a framework for Saga-pattern to solve distribute transaction problem.
// In saga-pattern, Saga is a long-lived transaction came up with many small sub-transaction.
// ExecutionCoordinator(SEC) is coordinator for sub-transactions execute and saga-log written.
// Sub-transaction is normal business operation, it contain a Action and action's Compensate.
// Saga-Log is used to record saga process, and SEC will use it to decide next step and how to recovery from error.
//
// There is a great speak for Saga-pattern at https://www.youtube.com/watch?v=xDuwrtwYHu8
package saga

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"reflect"
	"time"

	"github.com/itimofeev/go-saga/storage"
	"golang.org/x/net/context"
)

var Logger *logrus.Logger

func init() {
	Logger = logrus.New()
	Logger.Level = logrus.DebugLevel
}

func SetLogger(l *logrus.Logger) {
	Logger = l
}

// Saga presents current execute transaction.
// A Saga constituted by small sub-transactions.
type Saga struct {
	logID   string
	context context.Context
	sec     *ExecutionCoordinator
	storage storage.Storage
}

func (s *Saga) startSaga() {
	log := &storage.Log{
		SagaLogID: s.logID,
		Type:      storage.SagaStart,
		Time:      time.Now(),
	}
	err := s.storage.AppendLog(log)
	if err != nil {
		panic(err)
	}
}

// ExecSub executes a sub-transaction for given subTxID(which define in SEC initialize) and arguments.
// it returns current Saga.
func (s *Saga) ExecSub(subTxID string, args ...interface{}) *Saga {
	subTxDef := s.sec.mustFindSubTxDef(subTxID)
	log := &storage.Log{
		SagaLogID: s.logID,
		Type:      storage.ActionStart,
		SubTxID:   subTxID,
		Time:      time.Now(),
		Params:    MarshalParam(s.sec, args),
	}
	err := s.storage.AppendLog(log)
	if err != nil {
		panic(err)
	}

	params := make([]reflect.Value, 0, len(args)+1)
	params = append(params, reflect.ValueOf(s.context))
	for _, arg := range args {
		params = append(params, reflect.ValueOf(arg))
	}
	result := subTxDef.action.Call(params)
	if isReturnError(result) {
		s.Abort()
		return s
	}

	log = &storage.Log{
		SagaLogID: s.logID,
		Type:      storage.ActionEnd,
		SubTxID:   subTxID,
		Time:      time.Now(),
	}
	err = s.storage.AppendLog(log)
	if err != nil {
		panic(err)
	}
	return s
}

// EndSaga finishes a Saga's execution.
func (s *Saga) EndSaga() {
	log := &storage.Log{
		SagaLogID: s.logID,
		Type:      storage.SagaEnd,
		Time:      time.Now(),
	}
	err := s.storage.AppendLog(log)
	if err != nil {
		panic("Add log Failure")
	}
	err = s.storage.Cleanup(s.logID)
	if err != nil {
		panic(err)
	}
}

// Abort stop and compensate to rollback to start situation.
// This method will stop continue sub-transaction and do Compensate for executed sub-transaction.
// SubTx will call this method internal.
func (s *Saga) Abort() {
	logs, err := s.storage.Lookup(s.logID)
	if err != nil {
		panic(err)
	}
	log := &storage.Log{
		SagaLogID: s.logID,
		Type:      storage.SagaAbort,
		Time:      time.Now(),
	}
	err = s.storage.AppendLog(log)
	if err != nil {
		panic("Add log Failure")
	}
	fmt.Println(logs)
	// TODO !!!
	//for i := len(logs) - 1; i >= 0; i-- {
	//	logData := logs[i]
	//	log := mustUnmarshalLog(logData)
	//	if log.Type == ActionStart {
	//		if err := s.compensate(log); err != nil {
	//			panic("Compensate Failure..")
	//		}
	//	}
	//}
}

func (s *Saga) compensate(tlog *storage.Log) error {
	clog := &storage.Log{
		SagaLogID: s.logID,
		Type:      storage.CompensateStart,
		SubTxID:   tlog.SubTxID,
		Time:      time.Now(),
	}
	err := s.storage.AppendLog(clog)
	if err != nil {
		panic("Add log Failure")
	}

	args := UnmarshalParam(s.sec, tlog.Params)

	params := make([]reflect.Value, 0, len(args)+1)
	params = append(params, reflect.ValueOf(s.context))
	params = append(params, args...)

	subDef := s.sec.mustFindSubTxDef(tlog.SubTxID)
	result := subDef.compensate.Call(params)
	if isReturnError(result) {
		panic(result)
	}

	clog = &storage.Log{
		SagaLogID: s.logID,
		Type:      storage.CompensateEnd,
		SubTxID:   tlog.SubTxID,
		Time:      time.Now(),
	}
	err = s.storage.AppendLog(clog)
	if err != nil {
		panic("Add log Failure")
	}
	return nil
}

func isReturnError(result []reflect.Value) bool {
	if len(result) == 1 && !result[0].IsNil() {
		return true
	}
	return false
}
