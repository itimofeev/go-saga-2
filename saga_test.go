package saga_test

import (
	"fmt"
	"github.com/lysu/go-saga"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"testing"
)

var (
	storage saga.Storage
	sec     saga.ExecutionCoordinator
)

func initIt(mode FailureMode) {
	storage, _ = saga.NewMemStorage()

	sec = saga.NewSEC(storage)

	sec.AddSubTxDef("deduce", DeduceAccount, CompensateDeduce).
		AddSubTxDef("deposit", DepositAccount, CompensateDeposit).
		AddSubTxDef("test", PTest1, PTest1)

	memDB = map[string]int{
		"foo": 200,
		"bar": 0,
	}

	testMode = mode
}

func TestAllSuccess(t *testing.T) {

	initIt(OK)

	from, to := "foo", "bar"
	amount := 100

	ctx := context.Background()

	var sagaID uint64 = 1
	sec.StartSaga(ctx, sagaID).
		SubTx("deduce", from, amount).
		SubTx("deposit", to, amount).
		EndSaga()

	assert.Equal(t, 100, memDB[from])
	assert.Equal(t, 100, memDB[to])

	logs, err := storage.Lookup("saga_1")
	assert.NoError(t, err)
	assert.Equal(t, 6, len(logs))

}

func TestDepositFail(t *testing.T) {

	// initInStartup
	initIt(DepositFail)

	from, to := "foo", "bar"
	amount := 100

	ctx := context.Background()

	var sagaID uint64 = 1
	sec.StartSaga(ctx, sagaID).
		SubTx("deduce", from, amount).
		SubTx("deposit", to, amount).
		EndSaga()

	// assert
	assert.Equal(t, 200, memDB[from])
	assert.Equal(t, -100, memDB[to]) // BUG fix test

	logs, err := storage.Lookup("saga_1")
	assert.NoError(t, err)
	t.Logf("%v", logs)
	assert.Equal(t, 10, len(logs))

}

// This example show how to use init a SEC and add Sub-transaction to SEC, finally use a transfer operation as demo.
// deduce `100` from foo then deposit 100 into `bar`.
func ExampleTransfer() {

	// 1. Define sub-transaction method, anonymous method is NOT required, Just define them as normal way.

	DeduceAccount := func(ctx context.Context, account string, amount int) error {
		// Do deduce amount from account
		return nil
	}
	CompensateDeduce := func(ctx context.Context, account string, amount int) error {
		// Compensate deduce amount to account
		return nil
	}
	DepositAccount := func(ctx context.Context, account string, amount int) error {
		// Do deposit amount to account
		return nil
	}
	CompensateDeposit := func(ctx context.Context, account string, amount int) error {
		// Compensate deposit amount from account
		return nil
	}

	// 2. Init SEC as global SINGLETON(this demo not..), and add Sub-transaction definition into SEC.

	storage, _ := saga.NewMemStorage()
	sec := saga.NewSEC(storage)
	sec.AddSubTxDef("deduce", DeduceAccount, CompensateDeduce).
		AddSubTxDef("deposit", DepositAccount, CompensateDeposit)

	// 3. Start a saga to transfer 100 from foo to bar.

	from, to := "foo", "bar"
	amount := 100
	ctx := context.Background()

	var sagaID uint64 = 1
	sec.StartSaga(ctx, sagaID).
		SubTx("deduce", from, amount).
		SubTx("deposit", to, amount).
		EndSaga()

	// 4. done.
}

type FailureMode int

const (
	OK = iota
	DeduceFail
	DepositFail
)

var (
	memDB    map[string]int
	testMode FailureMode
)

func DeduceAccount(ctx context.Context, account string, amount int) error {
	if testMode == DeduceFail {
		return fmt.Errorf("Deduce failure")
	}
	memDB[account] = (memDB[account] - amount)
	return nil
}

func CompensateDeduce(ctx context.Context, account string, amount int) error {
	memDB[account] = (memDB[account] + amount)
	return nil
}

func DepositAccount(ctx context.Context, account string, amount int) error {
	if testMode == DepositFail {
		return fmt.Errorf("Deposit failure")
	}
	memDB[account] = (memDB[account] + amount)
	return nil
}

func CompensateDeposit(ctx context.Context, account string, amount int) error {
	memDB[account] = (memDB[account] - amount)
	return nil
}

func PTest1(ctx context.Context, name *string, age int) {

}
