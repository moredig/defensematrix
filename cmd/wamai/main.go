package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/moredig/defensematrix/internal/deception"
	"github.com/moredig/defensematrix/internal/detection"
	"github.com/moredig/defensematrix/internal/docker"
	"github.com/moredig/defensematrix/internal/fleet"
	"github.com/moredig/defensematrix/internal/proxy"
)

const (
	BoatImage       = "defensematrix-boat:latest"
	GatewayPort     = "8080" // External-facing port
	ScanThreshold   = 5      // Hits before auto-nuke
	RecycleInterval = 500    // Recycler check interval in ms
	BoatNetwork     = "defense-net"
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

	// Step 2 — Initialise fleet
	fmt.Println("[WAMAI] Initialising fleet...")
	f := fleet.NewFleet(dockerClient, BoatImage, BoatNetwork)
	if err := f.Launch(ctx); err != nil {
		fmt.Printf("[WAMAI] Fatal: %v\n", err)
		os.Exit(1)
	}

	// Step 3 — Drop breadcrumbs into frontline boat
	fmt.Println("[WAMAI] Dropping breadcrumbs...")
	frontline := f.GetFrontline()
	if frontline == nil {
		fmt.Println("[WAMAI] Fatal: fleet has no frontline boat")
		os.Exit(1)
	}
	files := make(map[string]string)
	for _, crumb := range deception.GenerateBreadcrumbs() {
		files[path.Join("boat-lures", crumb.Filename)] = crumb.Content
	}
	if err := dockerClient.CopyFiles(ctx, frontline.ID, "/tmp", files); err != nil {
		fmt.Printf("[WAMAI] Warning: breadcrumb drop failed: %v\n", err)
	}

	var gateway *proxy.Gateway

	// Step 4 — Initialise scanner with nuke callback
	fmt.Println("[WAMAI] Arming scanner...")
	scanner := detection.NewScanner(ScanThreshold, func() {
		fmt.Println("[WAMAI] Scanner triggered nuke — rotating fleet...")
		if err := f.Rotate(ctx); err != nil {
			fmt.Printf("[WAMAI] Rotation error: %v\n", err)
		}
	})

	// Step 5 — Start recycler
	fmt.Println("[WAMAI] Starting recycler...")
	recycler := fleet.NewRecycler(f, RecycleInterval)
	recycler.Start(ctx)

	// Step 6 — Start gateway
	fmt.Println("[WAMAI] Arming gateway...")
	gateway, err = proxy.NewGateway("http://"+frontline.Name, scanner)
	if err != nil {
		fmt.Printf("[WAMAI] Fatal: %v\n", err)
		os.Exit(1)
	}
	f.SetOnRotate(func() {
		if frontline := f.GetFrontline(); frontline != nil {
			if err := gateway.Retarget("http://" + frontline.Name); err != nil {
				fmt.Printf("[WAMAI] Retarget error: %v\n", err)
			}
		}
	})

	// Step 7 — Start multiplexer
	mux := proxy.NewMultiplexer(gateway, func() {
		fmt.Println("[WAMAI] Multiplexer triggered nuke — rotating fleet...")
		if err := f.Rotate(ctx); err != nil {
			fmt.Printf("[WAMAI] Rotation error: %v\n", err)
		}
	})
	_ = mux

	// Step 8 — Listen for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Step 9 — Start gateway in goroutine
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
