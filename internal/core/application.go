package core

import (
	"context"
	"strings"
	"time"
)

// Dependencies contains the foundational ports required by the V2 core.
type Dependencies struct {
	Clock        Clock
	IDs          IDGenerator
	Transactions TransactionManager
}

// Application is the persistence- and protocol-independent V2 application core.
type Application struct {
	clock        Clock
	ids          IDGenerator
	transactions TransactionManager
}

// NewApplication validates foundational ports and constructs the V2 core.
func NewApplication(dependencies Dependencies) (*Application, error) {
	const operation = "core.NewApplication"

	if dependencies.Clock == nil {
		return nil, missingDependency(operation, "clock")
	}
	if dependencies.IDs == nil {
		return nil, missingDependency(operation, "ids")
	}
	if dependencies.Transactions == nil {
		return nil, missingDependency(operation, "transactions")
	}

	return &Application{
		clock:        dependencies.Clock,
		ids:          dependencies.IDs,
		transactions: dependencies.Transactions,
	}, nil
}

// Now returns the application clock in UTC.
func (a *Application) Now() time.Time {
	if a == nil || a.clock == nil {
		return time.Time{}
	}
	return a.clock.Now().UTC()
}

// NewID asks the configured ID generator for an opaque non-empty identifier.
func (a *Application) NewID(ctx context.Context) (string, error) {
	const operation = "core.Application.NewID"

	if a == nil || a.ids == nil {
		return "", NewError(CodeInternal, operation, "application is not initialized")
	}
	if ctx == nil {
		return "", NewFieldError(CodeInvalidArgument, operation, "context", "context is required")
	}

	id, err := a.ids.NewID(ctx)
	if err != nil {
		return "", preserveApplicationError(operation, "generate id", err)
	}
	if strings.TrimSpace(id) == "" {
		return "", NewError(CodeInternal, operation, "id generator returned an empty id")
	}
	return id, nil
}

// WithinTransaction executes fn through the configured transaction adapter.
func (a *Application) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	const operation = "core.Application.WithinTransaction"

	if a == nil || a.transactions == nil {
		return NewError(CodeInternal, operation, "application is not initialized")
	}
	if ctx == nil {
		return NewFieldError(CodeInvalidArgument, operation, "context", "context is required")
	}
	if fn == nil {
		return NewFieldError(CodeInvalidArgument, operation, "callback", "callback is required")
	}

	return preserveApplicationError(
		operation,
		"transaction failed",
		a.transactions.WithinTransaction(ctx, fn),
	)
}
