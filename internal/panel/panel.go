package panel

import (
	"embed"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/moredig/defensematrix/internal/detection"
	"github.com/moredig/defensematrix/internal/fleet"
)

//go:embed dashboard.html
var dashboard embed.FS

type statusResponse struct {
	UpdatedAt  time.Time          `json:"updatedAt"`
	Boats      []fleet.BoatStatus `json:"boats"`
	Rotations  uint64             `json:"rotations"`
	TotalHits  int                `json:"totalHits"`
	RecentHits []hitResponse      `json:"recentHits"`
	Running    bool               `json:"running"`
}

type hitResponse struct {
	Pattern   string    `json:"pattern"`
	Payload   string    `json:"payload"`
	Level     string    `json:"level"`
	Timestamp time.Time `json:"timestamp"`
}

// NewHandler creates the control panel and its status and process controls.
func NewHandler(f *fleet.Fleet, scanner *detection.Scanner, onStop, onRestart, onShutdown func()) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/stop", processControl(onStop, "stopping"))
	mux.HandleFunc("/api/restart", processControl(onRestart, "restarting"))
	mux.HandleFunc("/api/shutdown", processControl(onShutdown, "shutting down"))
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		fleetStatus := f.Status()
		scannerStats := scanner.GetStats(20)
		response := statusResponse{
			UpdatedAt:  time.Now().UTC(),
			Boats:      fleetStatus.Boats,
			Rotations:  fleetStatus.Rotations,
			TotalHits:  scannerStats.TotalHits,
			RecentHits: make([]hitResponse, 0, len(scannerStats.RecentHits)),
			Running:    fleetStatus.Running,
		}
		for _, hit := range scannerStats.RecentHits {
			response.RecentHits = append(response.RecentHits, hitResponse{
				Pattern:   hit.Pattern,
				Payload:   hit.Payload,
				Level:     hit.Level.String(),
				Timestamp: hit.Timestamp,
			})
		}

		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			return
		}
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page, err := dashboard.ReadFile("dashboard.html")
		if err != nil {
			http.Error(w, "panel unavailable", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(page)
	})
	return mux
}

func processControl(callback func(), status string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !sameOrigin(r) {
			http.Error(w, "same-origin request required", http.StatusForbidden)
			return
		}
		if callback == nil {
			http.Error(w, "control unavailable", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": status})
		callback()
	}
}

func sameOrigin(r *http.Request) bool {
	origin, err := url.Parse(r.Header.Get("Origin"))
	return err == nil && origin.Scheme == "http" && origin.Host == r.Host && origin.User == nil && origin.Path == "" && origin.RawQuery == "" && origin.Fragment == ""
}
