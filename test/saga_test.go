package saga

import (
	"context"
	"errors"
	"github.com/itimofeev/go-saga"
	"github.com/itimofeev/go-saga/storage/memory"
	"github.com/stretchr/testify/require"
	"testing"
)

var fooAcc = 1000
var barAcc = 2000

var deduceFunc = func(ctx context.Context, account string, amount int) error {
	if account == "foo" {
		fooAcc -= amount
	} else if account == "bar" {
		barAcc -= amount
	} else {
		panic("unknown account")
	}
	return nil
}

var depositFunc = func(ctx context.Context, account string, amount int) error {
	if account == "foo" {
		fooAcc += amount
	} else if account == "bar" {
		barAcc += amount
	} else {
		panic("unknown account")
	}
	return nil
}

var errFunc = func(ctx context.Context, account string, amount int) error {
	return errors.New("hello")
}

func TestName(t *testing.T) {
	DeduceAccount := deduceFunc
	CompensateDeduce := depositFunc
	DepositAccount := depositFunc
	CompensateDeposit := deduceFunc

	saga := saga.NewSEC(memory.New())

	saga.AddSubTxDef("deduce", DeduceAccount, CompensateDeduce).
		AddSubTxDef("deposit", DepositAccount, CompensateDeposit)

	// 3. Start a saga to transfer 100 from foo to bar.

	from, to := "foo", "bar"
	amount := 100
	ctx := context.Background()

	var sagaID uint64 = 2
	saga.StartSaga(ctx, sagaID).
		ExecSub("deduce", from, amount).
		ExecSub("deposit", to, amount).
		EndSaga()

	require.Equal(t, fooAcc, 900)
	require.Equal(t, barAcc, 2100)
}

func TestName2(t *testing.T) {
	DeduceAccount := deduceFunc
	CompensateDeduce := depositFunc
	DepositAccount := errFunc
	CompensateDeposit := deduceFunc

	saga := saga.NewSEC(memory.New())

	saga.AddSubTxDef("deduce", DeduceAccount, CompensateDeduce).
		AddSubTxDef("deposit", DepositAccount, CompensateDeposit)

	// 3. Start a saga to transfer 100 from foo to bar.

	from, to := "foo", "bar"
	amount := 100
	ctx := context.Background()

	var sagaID uint64 = 2
	saga.StartSaga(ctx, sagaID).
		ExecSub("deduce", from, amount).
		ExecSub("deposit", to, amount).
		EndSaga()

	//require.Equal(t, 1000, fooAcc)
	require.Equal(t, 2000, barAcc)
}
