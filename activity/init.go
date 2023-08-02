package activity

import (
	"context"

	"go.olapie.com/ola/internal/delegate"
)

func init() {
	delegate.GetSession = func(ctx context.Context) any {
		return FromContext(ctx).Session
	}
}
