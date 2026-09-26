package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/moredig/defensematrix/internal/detection"
)

func TestGatewayScansAndForwardsRequestBody(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	defer backend.Close()

	scanner := detection.NewScanner(10, nil)
	gateway, err := NewGateway(backend.URL, scanner)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader("username=admin&value=hello"))
	recorder := httptest.NewRecorder()
	gateway.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK || recorder.Body.String() != "username=admin&value=hello" {
		t.Fatalf("unexpected response: %d %q", recorder.Code, recorder.Body.String())
	}
}

func TestGatewayBlocksAttackBeforeForwarding(t *testing.T) {
	backendRequests := 0
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backendRequests++
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	gateway, err := NewGateway(backend.URL, detection.NewScanner(10, nil))
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodGet, "/items?id=1%20UNION%20SELECT%20NULL,NULL", nil)
	response := httptest.NewRecorder()
	gateway.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected attack to be blocked, got %d", response.Code)
	}
	if backendRequests != 0 {
		t.Fatalf("blocked request reached backend %d times", backendRequests)
	}
}
