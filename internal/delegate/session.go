package delegate

import "context"

var GetSession func(ctx context.Context) any
