package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/moredig/defensematrix/internal/deception"
	"github.com/moredig/defensematrix/internal/detection"
	"github.com/moredig/defensematrix/internal/security"
)

// Gateway is the main reverse proxy and traffic interceptor
type Gateway struct {
	mu          sync.RWMutex
	targetURL   *url.URL
	proxy       *httputil.ReverseProxy
	wafHandler  http.Handler
	scanner     *detection.Scanner
	baitProfile *deception.BaitProfile
}

// NewGateway creates a new Gateway pointing at a backend boat
func NewGateway(targetAddr string, scanner *detection.Scanner) (*Gateway, error) {
	target, err := url.Parse(targetAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid target address: %w", err)
	}

	profiles := deception.Profiles()
	profile := profiles["apache-legacy"] // Default bait profile

	g := &Gateway{
		targetURL:   target,
		scanner:     scanner,
		baitProfile: profile,
	}

	g.proxy = &httputil.ReverseProxy{
		Director: g.director,
	}
	g.wafHandler, err = security.WrapHTTP(http.HandlerFunc(g.forwardRequest))
	if err != nil {
		return nil, fmt.Errorf("initialize HTTP request firewall: %w", err)
	}

	fmt.Printf("[WAMAI] Gateway armed — target: %s\n", targetAddr)
	return g, nil
}

// director modifies the incoming request before forwarding
func (g *Gateway) director(req *http.Request) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	req.URL.Scheme = g.targetURL.Scheme
	req.URL.Host = g.targetURL.Host
	req.Host = g.targetURL.Host

}

// ServeHTTP handles incoming requests — injects bait headers and proxies
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := g.scanRequest(r); err != nil {
		http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
		return
	}
	g.wafHandler.ServeHTTP(w, r)
}

func (g *Gateway) forwardRequest(w http.ResponseWriter, r *http.Request) {
	// Inject fake server identity headers
	deception.InjectHeaders(w, g.baitProfile)

	// Forward to the active boat
	g.proxy.ServeHTTP(w, r)
}

func (g *Gateway) scanRequest(r *http.Request) error {
	payload := r.URL.RequestURI()
	for key, values := range r.Header {
		payload += "\n" + key + ": " + fmt.Sprint(values)
	}

	if r.Body != nil {
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20+1))
		if err != nil {
			return err
		}
		if len(body) > 1<<20 {
			return fmt.Errorf("request body exceeds 1 MiB")
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		payload += "\n" + string(body)
	}

	g.scanner.Scan(payload)
	return nil
}

// Retarget switches the proxy to point at a new boat address
// Called during zero-downtime rotation
func (g *Gateway) Retarget(newAddr string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	target, err := url.Parse(newAddr)
	if err != nil {
		return fmt.Errorf("invalid retarget address: %w", err)
	}

	g.targetURL = target
	g.proxy = &httputil.ReverseProxy{
		Director: g.director,
	}

	fmt.Printf("[WAMAI] Gateway retargeted → %s\n", newAddr)
	return nil
}

// Listen starts the gateway on a given port
func (g *Gateway) Listen(port string) error {
	addr := ":" + port
	fmt.Printf("[WAMAI] Gateway listening on %s\n", addr)
	return http.ListenAndServe(addr, g)
}
