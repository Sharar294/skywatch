package service

import (
	"context"
	"encoding/json"
	"skywatch/internal/domain"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	client *redis.Client
}

func NewStore(addr string) *Store {
	return &Store{
		client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

// SaveLatestFlights converts the flight slice to JSON and saves it in Redis.
func (s *Store) SaveLatestFlights(ctx context.Context, flights []domain.Flight) error {
	data, err := json.Marshal(flights)
	if err != nil {
		return err
	}
	// We set the key "latest_flights" to expire in 60s.
	// If the worker crashes, the API won't show stale data forever.
	return s.client.Set(ctx, "latest_flights", data, 60*time.Second).Err()
}
