package bootstrap

import (
	"errors"
	"io"
	"sync"

	"github.com/mss-boot-io/meminfra/internal/core"
)

// Dependencies contains adapter implementations wired into the V2 core.
type Dependencies struct {
	Clock        core.Clock
	IDs          core.IDGenerator
	Transactions core.TransactionManager
	Closers      []io.Closer
}

// Runtime owns the composed application and adapter lifecycle.
type Runtime struct {
	Application *core.Application
	closers     []io.Closer
	closeOnce   sync.Once
	closeErr    error
}

// New constructs the application core and captures adapter resources for shutdown.
func New(dependencies Dependencies) (*Runtime, error) {
	application, err := core.NewApplication(core.Dependencies{
		Clock:        dependencies.Clock,
		IDs:          dependencies.IDs,
		Transactions: dependencies.Transactions,
	})
	if err != nil {
		return nil, err
	}

	closers := make([]io.Closer, 0, len(dependencies.Closers))
	for _, closer := range dependencies.Closers {
		if closer != nil {
			closers = append(closers, closer)
		}
	}

	return &Runtime{Application: application, closers: closers}, nil
}

// Close releases adapter resources in reverse construction order.
func (r *Runtime) Close() error {
	if r == nil {
		return nil
	}

	r.closeOnce.Do(func() {
		closeErrors := make([]error, 0, len(r.closers))
		for index := len(r.closers) - 1; index >= 0; index-- {
			if err := r.closers[index].Close(); err != nil {
				closeErrors = append(closeErrors, err)
			}
		}
		r.closeErr = errors.Join(closeErrors...)
	})
	return r.closeErr
}
