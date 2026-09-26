package fleet

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/moredig/defensematrix/internal/docker"
)

// Fleet manages the rolling trinity of Boats
type Fleet struct {
	mu        sync.Mutex
	boats     [3]*Boat
	docker    *docker.Client
	image     string // Docker image used for all boats
	network   string
	onRotate  func()
	rotations uint64
	stopped   bool
}

type BoatStatus struct {
	Name  string `json:"name"`
	Role  string `json:"role"`
	State string `json:"state"`
}

type Status struct {
	Boats     []BoatStatus `json:"boats"`
	Rotations uint64       `json:"rotations"`
	Running   bool         `json:"running"`
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

// Status returns a consistent, read-only snapshot of the fleet.
func (f *Fleet) Status() Status {
	f.mu.Lock()
	defer f.mu.Unlock()

	status := Status{
		Boats:     make([]BoatStatus, 0, len(f.boats)),
		Rotations: f.rotations,
		Running:   !f.stopped,
	}
	for _, boat := range f.boats {
		status.Boats = append(status.Boats, BoatStatus{
			Name:  boat.Name,
			Role:  boat.Role,
			State: boat.GetState().String(),
		})
	}
	return status
}

// Rotate executes the zero-downtime handoff
// Frontline → Graveyard, Shadow → Frontline, Graveyard → Shadow
func (f *Fleet) Rotate(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.stopped {
		return fmt.Errorf("fleet is shutting down")
	}

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
	f.rotations++

	fmt.Println("[WAMAI] Rotation complete.")
	if f.onRotate != nil {
		go f.onRotate()
	}
	return nil
}

// Stop gracefully stops all boats and prevents further fleet rotations.
func (f *Fleet) Stop(ctx context.Context) error {
	f.mu.Lock()
	if f.stopped {
		f.mu.Unlock()
		return nil
	}
	f.stopped = true
	boats := f.boats
	for _, boat := range boats {
		boat.SetState(StateStopped)
	}
	f.mu.Unlock()

	stopErrors := make(chan error, len(boats))
	var waitGroup sync.WaitGroup
	for _, boat := range boats {
		if boat.ID == "" {
			continue
		}
		waitGroup.Add(1)
		go func(boat *Boat) {
			defer waitGroup.Done()
			if err := f.docker.StopBoat(ctx, boat.ID); err != nil {
				stopErrors <- fmt.Errorf("stop %s: %w", boat.Name, err)
			}
		}(boat)
	}
	waitGroup.Wait()
	close(stopErrors)
	var joinedErrors []error
	for err := range stopErrors {
		joinedErrors = append(joinedErrors, err)
	}
	return errors.Join(joinedErrors...)
}

// Resume starts the existing boat containers and preserves their current roles.
func (f *Fleet) Resume(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.stopped {
		return nil
	}

	for _, boat := range f.boats {
		id, err := f.docker.StartBoat(ctx, f.image, boat.Name, f.network)
		if err != nil {
			resumeErr := fmt.Errorf("resume %s: %w", boat.Name, err)
			stopErrors := []error{resumeErr}
			for _, item := range f.boats {
				if stopErr := f.docker.StopBoat(ctx, item.ID); stopErr != nil {
					stopErrors = append(stopErrors, fmt.Errorf("rollback %s: %w", item.Name, stopErr))
				}
			}
			return errors.Join(stopErrors...)
		}
		boat.ID = id
		boat.SetState(StateIdle)
	}

	f.stopped = false
	fmt.Println("[WAMAI] Fleet resumed.")
	return nil
}
