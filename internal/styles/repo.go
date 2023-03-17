package styles

import "context"

type StyleRepo interface {
	Random(ctx context.Context) Style
	Get(ctx context.Context, number string) *Style
}
