package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/errdefs"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

// StartBoat spins up a new container from a given image, returns container ID
func (c *Client) StartBoat(ctx context.Context, image string, name string, networkName string) (string, error) {
	existing, err := c.cli.ContainerInspect(ctx, name)
	if err == nil {
		if !existing.State.Running {
			if err := c.cli.ContainerStart(ctx, existing.ID, types.ContainerStartOptions{}); err != nil {
				return "", fmt.Errorf("failed to start existing container %s: %w", name, err)
			}
		}
		return existing.ID, nil
	}

	var networkingConfig *network.NetworkingConfig
	if networkName != "" {
		networkingConfig = &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				networkName: {},
			},
		}
	}

	resp, err := c.cli.ContainerCreate(ctx,
		&container.Config{
			Image: image,
			Tty:   false,
		},
		&container.HostConfig{
			AutoRemove: false,
		},
		networkingConfig, (*specs.Platform)(nil), name,
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

// CopyFiles copies named files into a directory inside a running container.
func (c *Client) CopyFiles(ctx context.Context, containerID string, targetDir string, files map[string]string) error {
	var archive bytes.Buffer
	tarWriter := tar.NewWriter(&archive)
	for name, content := range files {
		header := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to archive %s: %w", name, err)
		}
		if _, err := tarWriter.Write([]byte(content)); err != nil {
			return fmt.Errorf("failed to archive %s: %w", name, err)
		}
	}
	if err := tarWriter.Close(); err != nil {
		return fmt.Errorf("failed to finalize container files: %w", err)
	}

	if err := c.cli.CopyToContainer(ctx, containerID, targetDir, &archive, types.CopyToContainerOptions{}); err != nil {
		return fmt.Errorf("failed to copy files into container %s: %w", containerID[:12], err)
	}
	return nil
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

// StopBoat gracefully stops a running boat without removing its container.
func (c *Client) StopBoat(ctx context.Context, containerID string) error {
	if containerID == "" {
		return nil
	}

	state, err := c.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to inspect boat %s: %w", containerID[:12], err)
	}
	if !state.State.Running {
		return nil
	}

	timeout := 5 * time.Second
	if err := c.cli.ContainerStop(ctx, containerID, &timeout); err != nil {
		if errdefs.IsNotModified(err) {
			return nil
		}
		return fmt.Errorf("failed to stop boat %s: %w", containerID[:12], err)
	}
	fmt.Printf("[WAMAI] Boat stopped: %s\n", containerID[:12])
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
