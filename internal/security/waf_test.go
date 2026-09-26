package security

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWAFBlocksCommonWebAttacksBeforeNextHandler(t *testing.T) {
	forwarded := false
	handler, err := WrapHTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwarded = true
		w.WriteHeader(http.StatusOK)
	}))
	if err != nil {
		t.Fatalf("initialize WAF: %v", err)
	}

	attacks := []string{
		"/items?id=1%20UNION%20SELECT%20NULL,NULL",
		"/search?q=%3Cscript%3Ealert(1)%3C/script%3E",
		"/download?file=../../etc/passwd",
		"/run?cmd=%24%28id%29",
	}
	for _, target := range attacks {
		forwarded = false
		request := httptest.NewRequest(http.MethodGet, target, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if response.Code != http.StatusForbidden {
			t.Errorf("expected attack %q to be blocked, got %d", target, response.Code)
		}
		if forwarded {
			t.Errorf("blocked attack %q reached the next handler", target)
		}
	}
}

func TestWAFAllowsOrdinaryRequest(t *testing.T) {
	handler, err := WrapHTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	if err != nil {
		t.Fatalf("initialize WAF: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/products?search=admin+tools", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected ordinary request to pass, got %d", response.Code)
	}
}

func TestWAFRejectsOversizedRequestBody(t *testing.T) {
	handler, err := WrapHTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("oversized request reached the next handler")
	}))
	if err != nil {
		t.Fatalf("initialize WAF: %v", err)
	}

	body := strings.NewReader(strings.Repeat("a", requestBodyLimit+1))
	request := httptest.NewRequest(http.MethodPost, "/submit", body)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected oversized body to be rejected, got %d", response.Code)
	}
}
