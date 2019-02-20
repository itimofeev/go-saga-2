package storage

import (
	"time"
)

// LogType present type flag for Log
type LogType int

const (
	// SagaStart flag saga stared log
	SagaStart LogType = iota + 1
	// SagaEnd flag saga ended log
	SagaEnd
	// SagaAbort flag saga aborted
	SagaAbort
	// ActionStart flag action start log
	ActionStart
	// ActionEnd flag action end log
	ActionEnd
	// CompensateStart flag compensate start log
	CompensateStart
	// CompensateEnd flag compensate end log
	CompensateEnd
)

// Log presents Saga Log.
// Saga Log used to log execute status for saga,
// and SEC use it to compensate and retry.
type Log struct {
	SagaLogID string      `json:"sagaLogId,omitempty"`
	Type      LogType     `json:"type,omitempty"`
	SubTxID   string      `json:"subTxID,omitempty"`
	Time      time.Time   `json:"time,omitempty"`
	Params    []ParamData `json:"params,omitempty"`
}

// ParamData presents sub-transaction input parameter data.
// This structure used to store and restore tx input data into log.
type ParamData struct {
	ParamType string      `json:"paramType,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}
