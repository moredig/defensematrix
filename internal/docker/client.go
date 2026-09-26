package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

// Client wraps the Docker daemon connection
type Client struct {
	cli *client.Client
}

// New creates and validates a connection to the local Docker daemon
func New() (*Client, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("Docker daemon is unavailable; start Docker Desktop or the Docker engine and retry: %w", err)
	}

	// Ping the daemon to confirm it's alive
	_, err = cli.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Docker daemon is unavailable; start Docker Desktop or the Docker engine and retry: %w", err)
	}

	fmt.Println("[WAMAI] Docker daemon connected.")
	return &Client{cli: cli}, nil
}

// GetCLI exposes the raw Docker client for other packages
func (c *Client) GetCLI() *client.Client {
	return c.cli
}
