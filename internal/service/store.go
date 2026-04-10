package service

import (
	"context"
	"encoding/json"
	"skywatch/internal/domain"

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

// -------------------------------------------------------------------------
// WORKER METHOD (Producer)
// -------------------------------------------------------------------------

// SaveLatestFlights converts the flight slice to JSON and saves it in Redis.
func (s *Store) SaveLatestFlights(ctx context.Context, flights []domain.Flight) error {
	data, err := json.Marshal(flights)
	if err != nil {
		return err
	}
	// Expire in 60s so the API doesn't serve stale data if the worker crashes
	return s.client.Set(ctx, "latest_flights", data, 60*1000000000).Err()
}

// -------------------------------------------------------------------------
// API METHOD (Consumer)
// -------------------------------------------------------------------------

// GetLatestFlights allows the API to pull what the worker saved
func (s *Store) GetLatestFlights(ctx context.Context) ([]domain.Flight, error) {
	val, err := s.client.Get(ctx, "latest_flights").Result()
	if err != nil {
		return nil, err
	}

	var flights []domain.Flight
	err = json.Unmarshal([]byte(val), &flights)
	if err != nil {
		return nil, err
	}

	return flights, nil
}
