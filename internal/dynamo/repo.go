package dynamo

import "context"

type BrewRepo interface {
	Get(ctx context.Context, id string) (*Brew, error)
	GetByUserID(ctx context.Context, userID string, createdAfter string) ([]Brew, error)
	Save(ctx context.Context, brew *Brew) error
	Delete(ctx context.Context, id string) error
}

type LeaderboardRepo interface {
	Get(ctx context.Context, userID string) (*LeaderboardEntry, error)
	GetAll(ctx context.Context) ([]LeaderboardEntry, error)
	Save(ctx context.Context, leaderboardEntry *LeaderboardEntry) error
}
