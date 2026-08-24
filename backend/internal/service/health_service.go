package service

import "context"

type Pinger interface {
	Ping(ctx context.Context) error
}

type HealthService struct {
	db Pinger
}

func NewHealthService(db Pinger) *HealthService {
	return &HealthService{db: db}
}

func (s *HealthService) Ready(ctx context.Context) error {
	return s.db.Ping(ctx)
}
