package bootstrap

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"
)

type testClock struct{}

func (testClock) Now() time.Time {
	return time.Unix(0, 0)
}

type testIDs struct{}

func (testIDs) NewID(context.Context) (string, error) {
	return "test-id", nil
}

type testTransactions struct{}

func (testTransactions) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type recordingCloser struct {
	name  string
	order *[]string
	err   error
}

func (c recordingCloser) Close() error {
	*c.order = append(*c.order, c.name)
	return c.err
}

func TestRuntimeClosesAdaptersInReverseOrder(t *testing.T) {
	order := make([]string, 0, 2)
	secondError := errors.New("second close failed")
	runtime, err := New(Dependencies{
		Clock:        testClock{},
		IDs:          testIDs{},
		Transactions: testTransactions{},
		Closers: []io.Closer{
			recordingCloser{name: "first", order: &order},
			recordingCloser{name: "second", order: &order, err: secondError},
		},
	})
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}

	if err := runtime.Close(); !errors.Is(err, secondError) {
		t.Fatalf("unexpected close error: %v", err)
	}
	if len(order) != 2 || order[0] != "second" || order[1] != "first" {
		t.Fatalf("unexpected close order: %#v", order)
	}
	if err := runtime.Close(); !errors.Is(err, secondError) {
		t.Fatalf("unexpected repeated close error: %v", err)
	}
	if len(order) != 2 {
		t.Fatalf("runtime closed adapters more than once: %#v", order)
	}
}
