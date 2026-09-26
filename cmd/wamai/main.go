package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/moredig/defensematrix/internal/deception"
	"github.com/moredig/defensematrix/internal/detection"
	"github.com/moredig/defensematrix/internal/docker"
	"github.com/moredig/defensematrix/internal/fleet"
	"github.com/moredig/defensematrix/internal/panel"
	"github.com/moredig/defensematrix/internal/proxy"
)

const (
	BoatImage       = "defensematrix-boat:latest"
	GatewayPort     = "8080" // External-facing port
	PanelPort       = "9090"
	ScanThreshold   = 5   // Hits before auto-nuke
	RecycleInterval = 500 // Recycler check interval in ms
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

	// Step 5 — Initialise gateway
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

	// Step 6 — Start multiplexer
	mux := proxy.NewMultiplexer(gateway, func() {
		fmt.Println("[WAMAI] Multiplexer triggered nuke — rotating fleet...")
		if err := f.Rotate(ctx); err != nil {
			fmt.Printf("[WAMAI] Rotation error: %v\n", err)
		}
	})
	_ = mux

	// Step 7 — Prepare control and shutdown channels
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	stopRequests := make(chan struct{}, 1)
	restartRequests := make(chan struct{}, 1)
	shutdownRequests := make(chan struct{}, 1)
	serverErrors := make(chan error, 2)

	var recycler *fleet.Recycler
	startRecycler := func() {
		fmt.Println("[WAMAI] Starting recycler...")
		recycler = fleet.NewRecycler(f, RecycleInterval)
		recycler.Start(ctx)
	}
	stopRecycler := func() {
		if recycler != nil {
			recycler.Stop()
			recycler = nil
		}
	}

	var gatewayServer *http.Server
	startGateway := func() error {
		addr := ":" + GatewayPort
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("listen on %s: %w", addr, err)
		}
		server := &http.Server{
			Addr:              addr,
			Handler:           gateway,
			ReadHeaderTimeout: 5 * time.Second,
		}
		gatewayServer = server
		go func() {
			fmt.Printf("[WAMAI] Gateway listening on %s\n", addr)
			if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
				select {
				case serverErrors <- fmt.Errorf("gateway: %w", err):
				default:
				}
			}
		}()
		return nil
	}
	stopGateway := func(shutdownCtx context.Context) error {
		if gatewayServer == nil {
			return nil
		}
		server := gatewayServer
		gatewayServer = nil
		return server.Shutdown(shutdownCtx)
	}

	if err := startGateway(); err != nil {
		fmt.Printf("[WAMAI] Fatal: %v\n", err)
		return
	}
	startRecycler()

	panelAddr := os.Getenv("WAMAI_PANEL_ADDR")
	if panelAddr == "" {
		panelAddr = "127.0.0.1:" + PanelPort
	}
	gatewayHost := os.Getenv("WAMAI_BIND_ADDRESS")
	if gatewayHost == "" || gatewayHost == "0.0.0.0" {
		gatewayHost = "localhost"
	}
	panelServer := &http.Server{
		Addr: panelAddr,
		Handler: panel.NewHandler(f, scanner, func() {
			select {
			case stopRequests <- struct{}{}:
			default:
			}
		}, func() {
			select {
			case restartRequests <- struct{}{}:
			default:
			}
		}, func() {
			select {
			case shutdownRequests <- struct{}{}:
			default:
			}
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err := panelServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			select {
			case serverErrors <- fmt.Errorf("control panel: %w", err):
			default:
			}
		}
	}()

	fmt.Println("[WAMAI] Defense Matrix live. Press Ctrl+C to shut down.")
	fmt.Printf("[WAMAI] Bait site: http://%s:%s\n", gatewayHost, GatewayPort)
	fmt.Printf("[WAMAI] Control panel: http://localhost:%s\n", PanelPort)

	fullShutdown := false
	for !fullShutdown {
		select {
		case <-sigCh:
			fullShutdown = true
		case <-shutdownRequests:
			fullShutdown = true
		case err := <-serverErrors:
			fmt.Printf("[WAMAI] Server error: %v\n", err)
			fullShutdown = true
		case <-stopRequests:
			if !f.Status().Running {
				continue
			}
			fmt.Println("[WAMAI] Pausing gateway and fleet; control panel remains available.")
			pauseCtx, pauseCancel := context.WithTimeout(context.Background(), 20*time.Second)
			stopRecycler()
			if err := stopGateway(pauseCtx); err != nil {
				fmt.Printf("[WAMAI] Gateway pause error: %v\n", err)
			}
			if err := f.Stop(pauseCtx); err != nil {
				fmt.Printf("[WAMAI] Fleet pause error: %v\n", err)
			}
			pauseCancel()
			fmt.Println("[WAMAI] Wamai paused.")
		case <-restartRequests:
			if f.Status().Running {
				continue
			}
			fmt.Println("[WAMAI] Restarting gateway and fleet...")
			restartCtx, restartCancel := context.WithTimeout(context.Background(), 20*time.Second)
			var restartErr error
			if err := f.Resume(restartCtx); err != nil {
				restartErr = err
			}
			if restartErr == nil {
				frontline := f.GetFrontline()
				if frontline == nil {
					restartErr = fmt.Errorf("fleet has no frontline boat")
				} else if err := gateway.Retarget("http://" + frontline.Name); err != nil {
					restartErr = err
				}
			}
			if restartErr == nil {
				restartErr = startGateway()
			}
			if restartErr != nil {
				fmt.Printf("[WAMAI] Restart failed: %v\n", restartErr)
				if err := f.Stop(restartCtx); err != nil {
					fmt.Printf("[WAMAI] Restart rollback error: %v\n", err)
				}
			} else {
				startRecycler()
				fmt.Println("[WAMAI] Wamai restarted.")
			}
			restartCancel()
		}
	}

	fmt.Println("\n[WAMAI] Full shutdown — standing down.")
	stopRecycler()
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer shutdownCancel()
	if err := stopGateway(shutdownCtx); err != nil {
		fmt.Printf("[WAMAI] Gateway shutdown error: %v\n", err)
	}
	if err := panelServer.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("[WAMAI] Control panel shutdown error: %v\n", err)
	}
	if err := f.Stop(shutdownCtx); err != nil {
		fmt.Printf("[WAMAI] Fleet shutdown error: %v\n", err)
	}
	fmt.Println("[WAMAI] Defense Matrix offline.")
}
