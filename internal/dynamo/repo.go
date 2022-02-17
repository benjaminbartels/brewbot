package dynamo

import "context"

type BrewRepo interface {
	Get(ctx context.Context, id string) (*Brew, error)
	GetAll(ctx context.Context) ([]Brew, error)
	GetByUserID(ctx context.Context, userID string) ([]Brew, error)
	Save(ctx context.Context, brew *Brew) error
	Delete(ctx context.Context, id string) error
}
