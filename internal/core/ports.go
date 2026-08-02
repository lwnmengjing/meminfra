package core

import (
	"context"
	"time"
)

// Clock supplies time to core use cases. Implementations must return a valid instant.
type Clock interface {
	Now() time.Time
}

// IDGenerator supplies opaque identifiers to core use cases.
type IDGenerator interface {
	NewID(context.Context) (string, error)
}

// TransactionManager runs a callback in one adapter-owned transaction boundary.
type TransactionManager interface {
	WithinTransaction(context.Context, func(context.Context) error) error
}
