// ctxzerolog provides the Context idiom for embedding zerolog into the execution context
package ctxzerolog

import (
	"context"

	"github.com/rs/zerolog"
)

type ctxKey struct{}

var key ctxKey

type wrapper struct {
	logger zerolog.Logger
}

func New(ctx context.Context, log zerolog.Logger) context.Context {
	return context.WithValue(ctx, key, &wrapper{log})
}

func Set(ctx context.Context, changes zerolog.Context) {
	l, ok := ctx.Value(key).(*wrapper)
	if !ok || l == nil {
		return
	}
	l.logger = changes.Logger()
}

func Get(ctx context.Context) zerolog.Context {
	l, ok := ctx.Value(key).(*wrapper)
	if !ok || l == nil {
		return zerolog.Nop().With()
	}
	return l.logger.With()
}
