package panel

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/moredig/defensematrix/internal/detection"
	"github.com/moredig/defensematrix/internal/fleet"
)

func TestHandlerServesPanelAndLiveStatus(t *testing.T) {
	handler := NewHandler(fleet.NewFleet(nil, "", ""), detection.NewScanner(5, nil), nil, nil, nil)

	page := httptest.NewRecorder()
	handler.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/", nil))
	if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), "Fleet status") {
		t.Fatalf("unexpected panel response: status=%d", page.Code)
	}

	status := httptest.NewRecorder()
	handler.ServeHTTP(status, httptest.NewRequest(http.MethodGet, "/api/status", nil))
	if status.Code != http.StatusOK {
		t.Fatalf("unexpected status response: %d", status.Code)
	}
	var payload statusResponse
	if err := json.Unmarshal(status.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode status response: %v", err)
	}
	if len(payload.Boats) != 3 || payload.Boats[0].Role != "frontline" {
		t.Fatalf("unexpected fleet snapshot: %#v", payload.Boats)
	}
}

func TestStatusEndpointAllowsOnlyGet(t *testing.T) {
	handler := NewHandler(fleet.NewFleet(nil, "", ""), detection.NewScanner(5, nil), nil, nil, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/status", nil))
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected method not allowed, got %d", response.Code)
	}
}

func TestControlEndpointsRequireSameOriginPost(t *testing.T) {
	stopCalled := false
	restartCalled := false
	shutdownCalled := false
	handler := NewHandler(fleet.NewFleet(nil, "", ""), detection.NewScanner(5, nil), func() {
		stopCalled = true
	}, func() {
		restartCalled = true
	}, func() {
		shutdownCalled = true
	})

	crossOrigin := httptest.NewRequest(http.MethodPost, "http://localhost/api/stop", nil)
	crossOrigin.Header.Set("Origin", "http://attacker.example")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, crossOrigin)
	if response.Code != http.StatusForbidden || stopCalled {
		t.Fatalf("cross-origin stop was not rejected: status=%d called=%v", response.Code, stopCalled)
	}

	sameOrigin := httptest.NewRequest(http.MethodPost, "http://localhost/api/stop", nil)
	sameOrigin.Header.Set("Origin", "http://localhost")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, sameOrigin)
	if response.Code != http.StatusAccepted || !stopCalled {
		t.Fatalf("same-origin stop was not accepted: status=%d called=%v", response.Code, stopCalled)
	}

	restart := httptest.NewRequest(http.MethodPost, "http://localhost/api/restart", nil)
	restart.Header.Set("Origin", "http://localhost")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, restart)
	if response.Code != http.StatusAccepted || !restartCalled {
		t.Fatalf("same-origin restart was not accepted: status=%d called=%v", response.Code, restartCalled)
	}

	shutdown := httptest.NewRequest(http.MethodPost, "http://localhost/api/shutdown", nil)
	shutdown.Header.Set("Origin", "http://localhost")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, shutdown)
	if response.Code != http.StatusAccepted || !shutdownCalled {
		t.Fatalf("same-origin full stop was not accepted: status=%d called=%v", response.Code, shutdownCalled)
	}
}
