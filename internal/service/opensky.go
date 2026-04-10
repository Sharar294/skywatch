package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"skywatch/internal/domain"
)

// OpenSkyClient handles communication with the OpenSky Network API.
type OpenSkyClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewOpenSkyClient initializes a client with a timeout—crucial for SRE!
func NewOpenSkyClient() *OpenSkyClient {
	return &OpenSkyClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second, // Don't let a slow API hang your worker
		},
		baseURL: "https://opensky-network.org/api/states/all",
	}
}

// FetchFlights pulls the latest state vectors and maps them to our domain model.
func (c *OpenSkyClient) FetchFlights() ([]domain.Flight, error) {
	resp, err := c.httpClient.Get(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to reach OpenSky: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var data OpenSkyResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Use the mapper logic we wrote in Week 1
	return MapToFlights(data.States), nil
}