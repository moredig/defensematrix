package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/moredig/defensematrix/internal/deception"
	"github.com/moredig/defensematrix/internal/detection"
	"github.com/moredig/defensematrix/internal/docker"
	"github.com/moredig/defensematrix/internal/fleet"
	"github.com/moredig/defensematrix/internal/proxy"
)

const (
	BoatImage       = "ubuntu:22.04" // Base image for all boats
	GatewayPort     = "8080"         // External-facing port
	ScanThreshold   = 5              // Hits before auto-nuke
	RecycleInterval = 500            // Recycler check interval in ms
)

func main() {
	fmt.Println(`
██╗    ██╗ █████╗ ███╗   ███╗ █████╗ ██╗
██║    ██║██╔══██╗████╗ ████║██╔══██╗██║
██║ █╗ ██║███████║██╔████╔██║███████║██║
██║███╗██║██╔══██║██║╚██╔╝██║██╔══██║██║
╚███╔███╔╝██║  ██║██║ ╚═╝ ██║██║  ██║██║
 ╚══╝╚══╝ ╚═╝  ╚═╝╚═╝     ╚═╝╚═╝  ╚═╝╚═╝
Defense Matrix — Online.
	`)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Step 1 — Connect to Docker daemon
	fmt.Println("[WAMAI] Connecting to Docker daemon...")
	dockerClient, err := docker.New()
	if err != nil {
		fmt.Printf("[WAMAI] Fatal: %v\n", err)
		os.Exit(1)
	}

	// Step 2 — Pull boat image
	fmt.Printf("[WAMAI] Pulling boat image: %s\n", BoatImage)
	if err := dockerClient.PullImage(ctx, BoatImage); err != nil {
		fmt.Printf("[WAMAI] Fatal: %v\n", err)
		os.Exit(1)
	}

	// Step 3 — Initialise fleet
	fmt.Println("[WAMAI] Initialising fleet...")
	f := fleet.NewFleet(dockerClient, BoatImage)
	if err := f.Launch(ctx); err != nil {
		fmt.Printf("[WAMAI] Fatal: %v\n", err)
		os.Exit(1)
	}

	// Step 4 — Drop breadcrumbs into frontline boat
	fmt.Println("[WAMAI] Dropping breadcrumbs...")
	if err := deception.DropBreadcrumbs("/tmp/boat-lures"); err != nil {
		fmt.Printf("[WAMAI] Warning: breadcrumb drop failed: %v\n", err)
	}

	// Step 5 — Initialise scanner with nuke callback
	fmt.Println("[WAMAI] Arming scanner...")
	scanner := detection.NewScanner(ScanThreshold, func() {
		fmt.Println("[WAMAI] Scanner triggered nuke — rotating fleet...")
		if err := f.Rotate(ctx); err != nil {
			fmt.Printf("[WAMAI] Rotation error: %v\n", err)
		}
	})

	// Step 6 — Start recycler
	fmt.Println("[WAMAI] Starting recycler...")
	recycler := fleet.NewRecycler(f, RecycleInterval)
	recycler.Start(ctx)

	// Step 7 — Start gateway
	fmt.Println("[WAMAI] Arming gateway...")
	gateway, err := proxy.NewGateway("http://localhost:9000", scanner)
	if err != nil {
		fmt.Printf("[WAMAI] Fatal: %v\n", err)
		os.Exit(1)
	}

	// Step 8 — Start multiplexer
	mux := proxy.NewMultiplexer(gateway, func() {
		fmt.Println("[WAMAI] Multiplexer triggered nuke — rotating fleet...")
		if err := f.Rotate(ctx); err != nil {
			fmt.Printf("[WAMAI] Rotation error: %v\n", err)
		}
	})
	_ = mux

	// Step 9 — Listen for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Step 10 — Start gateway in goroutine
	go func() {
		if err := gateway.Listen(GatewayPort); err != nil {
			fmt.Printf("[WAMAI] Gateway error: %v\n", err)
			cancel()
		}
	}()

	fmt.Printf("[WAMAI] Defense Matrix live on port %s. Press Ctrl+C to shut down.\n", GatewayPort)

	// Block until shutdown
	<-sigCh
	fmt.Println("\n[WAMAI] Shutdown signal received — standing down.")
	recycler.Stop()
	cancel()
	fmt.Println("[WAMAI] Defense Matrix offline.")
}
