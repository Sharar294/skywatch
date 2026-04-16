package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2/clientcredentials"
)

func main() {
	// Load variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it, relying on system environment variables")
	}

	clientID := os.Getenv("OPENSKY_CLIENT_ID")
	clientSecret := os.Getenv("OPENSKY_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("OPENSKY_CLIENT_ID and OPENSKY_CLIENT_SECRET must be set in your .env file or environment")
	}

	url := "https://opensky-network.org/api/states/all"
	ctx := context.Background()

	// Configure the standard OAuth2 Client Credentials flow
	// This natively uses the golang.org/x/oauth2 dependency
	conf := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		// OpenSky's official Keycloak OAuth token endpoint
		TokenURL: "https://auth.opensky-network.org/auth/realms/opensky-network/protocol/openid-connect/token",
	}

	// This initializes an HTTP client that automatically handles the token exchange
	client := conf.Client(ctx)

	client.Timeout = 10 * time.Second

	// Create a new request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Fetching data from OpenSky using oauth2 clientcredentials flow...")

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error executing request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected status code: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	// Print the returned metadata (HTTP Headers)
	fmt.Println("\n--- Response Metadata (HTTP Headers) ---")
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Printf("%s: %s\n", key, value)
		}
	}
	fmt.Println("----------------------------------------\n")

	// Read a small portion of the response to verify it worked
	body, err := io.ReadAll(io.LimitReader(resp.Body, 500))
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Success! First 500 bytes of response:\n%s...\n", string(body))
}
