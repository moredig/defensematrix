package fleet

import (
	"context"
	"fmt"
	"sync"

	"github.com/moredig/defensematrix/internal/docker"
)

// Fleet manages the rolling trinity of Boats
type Fleet struct {
	mu       sync.Mutex
	boats    [3]*Boat
	docker   *docker.Client
	image    string // Docker image used for all boats
	network  string
	onRotate func()
}

// SetOnRotate registers a callback invoked after a successful rotation.
func (f *Fleet) SetOnRotate(callback func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.onRotate = callback
}

// NewFleet initialises the three-boat rolling trinity
func NewFleet(dockerClient *docker.Client, image string, networkName string) *Fleet {
	return &Fleet{
		docker:  dockerClient,
		image:   image,
		network: networkName,
		boats: [3]*Boat{
			NewBoat("boat-1", "frontline"),
			NewBoat("boat-2", "shadow"),
			NewBoat("boat-3", "graveyard"),
		},
	}
}

// Launch spins up all three boats in Docker
func (f *Fleet) Launch(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	fmt.Println("[WAMAI] Launching fleet...")

	for _, boat := range f.boats {
		id, err := f.docker.StartBoat(ctx, f.image, boat.Name, f.network)
		if err != nil {
			return fmt.Errorf("failed to launch %s: %w", boat.Name, err)
		}
		boat.ID = id
	}

	fmt.Println("[WAMAI] Fleet is live.")
	return nil
}

// GetFrontline returns the currently active boat
func (f *Fleet) GetFrontline() *Boat {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, boat := range f.boats {
		if boat.Role == "frontline" {
			return boat
		}
	}
	return nil
}

// GetShadow returns the hot standby boat
func (f *Fleet) GetShadow() *Boat {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, boat := range f.boats {
		if boat.Role == "shadow" {
			return boat
		}
	}
	return nil
}

// Rotate executes the zero-downtime handoff
// Frontline → Graveyard, Shadow → Frontline, Graveyard → Shadow
func (f *Fleet) Rotate(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	fmt.Println("[WAMAI] Rotation triggered — nuking frontline...")

	var frontline, shadow, graveyard *Boat
	for _, boat := range f.boats {
		switch boat.Role {
		case "frontline":
			frontline = boat
		case "shadow":
			shadow = boat
		case "graveyard":
			graveyard = boat
		}
	}

	if frontline == nil || shadow == nil || graveyard == nil {
		return fmt.Errorf("fleet integrity error: missing boat role")
	}

	// Step 1 — Nuke the frontline
	frontline.SetState(StateNuked)
	if err := f.docker.NukeBoat(ctx, frontline.ID); err != nil {
		return fmt.Errorf("nuke failed: %w", err)
	}

	// Step 2 — Promote shadow to frontline
	shadow.Promote()

	// Step 3 — Recycle graveyard into new shadow
	graveyard.Role = "shadow"
	id, err := f.docker.StartBoat(ctx, f.image, graveyard.Name, f.network)
	if err != nil {
		return fmt.Errorf("recycle failed: %w", err)
	}
	graveyard.ID = id
	graveyard.SetState(StateIdle)

	// Step 4 — Old frontline becomes new graveyard
	frontline.Demote()

	fmt.Println("[WAMAI] Rotation complete.")
	if f.onRotate != nil {
		go f.onRotate()
	}
	return nil
}
