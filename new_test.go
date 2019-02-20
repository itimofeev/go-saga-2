package saga

import (
	"context"
	"errors"
	"fmt"
	"github.com/itimofeev/go-saga/storage/postgres"
	"math/rand"
	"testing"
	"time"
)

var firstCount = 0
var secondCount = 0
var thirdCount = 0

var first = func(ctx context.Context, count *int) error {
	*count++
	return nil
}
var firstComp = func(ctx context.Context, count *int) error {
	*count--
	return nil
}

var errFunc123 = func(ctx context.Context, _ *int) error {
	return errors.New("hello")
}

func TestName3(t *testing.T) {
	storage := postgres.New()
	//storage := memory.New()
	saga := NewSEC(storage)

	saga.AddSubTxDef("deduce1", first, firstComp).
		AddSubTxDef("deduce2", errFunc123, firstComp).
		AddSubTxDef("deduce3", first, firstComp)

	// 3. Start a saga to transfer 100 from foo to bar.

	ctx := context.Background()

	rand.Seed(time.Now().Unix())
	sagaID := fmt.Sprintf("%d", rand.Int())

	ss := saga.StartSaga(ctx, sagaID)

	ss.ExecSub("deduce1", &firstCount)
	ss.ExecSub("deduce2", &secondCount)
	ss.ExecSub("deduce3", &thirdCount)
	ss.EndSaga()

	fmt.Println(firstCount, secondCount, thirdCount)

	//require.Equal(t, 1, firstCount)
	//require.Equal(t, 1, secondCount)
	//require.Equal(t, 1, thirdCount)

}
