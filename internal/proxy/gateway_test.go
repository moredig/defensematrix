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
