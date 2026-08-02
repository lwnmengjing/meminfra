package core

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeClock struct {
	now time.Time
}

func (f fakeClock) Now() time.Time {
	return f.now
}

type fakeIDGenerator struct {
	id  string
	err error
}

func (f fakeIDGenerator) NewID(context.Context) (string, error) {
	return f.id, f.err
}

type fakeTransactionManager struct {
	calls int
	err   error
}

func (f *fakeTransactionManager) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	f.calls++
	if f.err != nil {
		return f.err
	}
	return fn(ctx)
}

func TestNewApplicationRequiresFoundationalPorts(t *testing.T) {
	_, err := NewApplication(Dependencies{})
	if !IsErrorCode(err, CodeInvalidArgument) {
		t.Fatalf("expected invalid argument, got %v", err)
	}

	applicationError := new(Error)
	if !errors.As(err, &applicationError) || applicationError.Field != "clock" {
		t.Fatalf("unexpected error: %#v", err)
	}
}

func TestApplicationUsesPorts(t *testing.T) {
	instant := time.Date(2026, 8, 2, 12, 0, 0, 0, time.FixedZone("test", 8*60*60))
	transactions := &fakeTransactionManager{}
	application, err := NewApplication(Dependencies{
		Clock:        fakeClock{now: instant},
		IDs:          fakeIDGenerator{id: "evd_test"},
		Transactions: transactions,
	})
	if err != nil {
		t.Fatalf("new application: %v", err)
	}

	if got := application.Now(); !got.Equal(instant) || got.Location() != time.UTC {
		t.Fatalf("unexpected application time: %s", got)
	}

	id, err := application.NewID(context.Background())
	if err != nil {
		t.Fatalf("new id: %v", err)
	}
	if id != "evd_test" {
		t.Fatalf("unexpected id: %q", id)
	}

	called := false
	if err := application.WithinTransaction(context.Background(), func(context.Context) error {
		called = true
		return nil
	}); err != nil {
		t.Fatalf("within transaction: %v", err)
	}
	if !called || transactions.calls != 1 {
		t.Fatalf("unexpected transaction invocation: called=%v calls=%d", called, transactions.calls)
	}
}

func TestApplicationDoesNotReclassifyTypedErrors(t *testing.T) {
	expected := NewError(CodeConflict, "append.evidence", "dedupe conflict")
	application, err := NewApplication(Dependencies{
		Clock:        fakeClock{now: time.Now()},
		IDs:          fakeIDGenerator{err: expected},
		Transactions: &fakeTransactionManager{},
	})
	if err != nil {
		t.Fatalf("new application: %v", err)
	}

	_, err = application.NewID(context.Background())
	if !errors.Is(err, expected) {
		t.Fatalf("expected typed error to be preserved: %v", err)
	}
	if !IsErrorCode(err, CodeConflict) {
		t.Fatalf("unexpected code: %v", err)
	}
}
