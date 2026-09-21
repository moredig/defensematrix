package docker

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

// StartBoat spins up a new container from a given image, returns container ID
func (c *Client) StartBoat(ctx context.Context, image string, name string) (string, error) {
	resp, err := c.cli.ContainerCreate(ctx,
		&container.Config{
			Image: image,
			Tty:   false,
		},
		&container.HostConfig{
			AutoRemove: false,
		},
		nil, nil, name,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create container %s: %w", name, err)
	}

	if err := c.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container %s: %w", name, err)
	}

	fmt.Printf("[WAMAI] Boat launched: %s (%s)\n", name, resp.ID[:12])
	return resp.ID, nil
}

// NukeBoat forcefully stops and removes a container instantly
func (c *Client) NukeBoat(ctx context.Context, containerID string) error {
	timeout := 0 * time.Second
	if err := c.cli.ContainerStop(ctx, containerID, &timeout); err != nil {
		return fmt.Errorf("failed to stop container %s: %w", containerID[:12], err)
	}

	if err := c.cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}); err != nil {
		return fmt.Errorf("failed to remove container %s: %w", containerID[:12], err)
	}

	fmt.Printf("[WAMAI] Boat nuked: %s\n", containerID[:12])
	return nil
}

// PullImage pulls a Docker image if not already present
func (c *Client) PullImage(ctx context.Context, image string) error {
	out, err := c.cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", image, err)
	}
	defer out.Close()
	io.Copy(io.Discard, out)

	fmt.Printf("[WAMAI] Image ready: %s\n", image)
	return nil
}
