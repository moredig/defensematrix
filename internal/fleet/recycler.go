package fleet

import (
	"context"
	"fmt"
	"time"
)

// Recycler monitors the fleet and automatically triggers rotation
// when a boat is flagged as nuked
type Recycler struct {
	fleet    *Fleet
	interval time.Duration // How often to check boat states
	stopCh   chan struct{}
}

// NewRecycler creates a new Recycler for a given fleet
func NewRecycler(fleet *Fleet, intervalMs int) *Recycler {
	return &Recycler{
		fleet:    fleet,
		interval: time.Duration(intervalMs) * time.Millisecond,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the recycler loop in a goroutine
func (r *Recycler) Start(ctx context.Context) {
	go func() {
		fmt.Println("[WAMAI] Recycler online.")
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				r.check(ctx)
			case <-r.stopCh:
				fmt.Println("[WAMAI] Recycler stopped.")
				return
			case <-ctx.Done():
				fmt.Println("[WAMAI] Recycler context cancelled.")
				return
			}
		}
	}()
}

// Stop shuts down the recycler loop
func (r *Recycler) Stop() {
	close(r.stopCh)
}

// check inspects the frontline boat and triggers rotation if nuked
func (r *Recycler) check(ctx context.Context) {
	frontline := r.fleet.GetFrontline()
	if frontline == nil {
		fmt.Println("[WAMAI] Recycler: no frontline found.")
		return
	}

	if frontline.GetState() == StateNuked {
		fmt.Printf("[WAMAI] Recycler: %s is nuked — initiating rotation.\n", frontline.Name)

		if err := r.fleet.Rotate(ctx); err != nil {
			fmt.Printf("[WAMAI] Rotation error: %v\n", err)
			return
		}

		fmt.Println("[WAMAI] Recycler: rotation successful.")
	}
}
