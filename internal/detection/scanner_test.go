package detection

import (
	"testing"
	"time"
)

func TestScannerScansAndCopiesHits(t *testing.T) {
	triggered := make(chan struct{}, 1)
	scanner := NewScanner(2, func() { triggered <- struct{}{} })

	scanner.Scan("<script>alert(1)</script>")
	hits := scanner.GetHits()
	if len(hits) != 1 || hits[0].Pattern != "XSS Attempt" {
		t.Fatalf("unexpected hits: %#v", hits)
	}

	hits[0].Payload = "changed"
	if scanner.GetHits()[0].Payload == "changed" {
		t.Fatal("GetHits returned mutable scanner state")
	}

	scanner.Scan("<script>alert(2)</script>")
	select {
	case <-triggered:
	case <-time.After(time.Second):
		t.Fatal("expected threshold callback")
	}
}

func TestScannerClampsInvalidThreshold(t *testing.T) {
	triggered := make(chan struct{}, 1)
	scanner := NewScanner(0, func() { triggered <- struct{}{} })
	scanner.Scan("<script>alert(1)</script>")
	select {
	case <-triggered:
	case <-time.After(time.Second):
		t.Fatal("expected zero threshold to clamp to one")
	}
}

func TestScannerStatsLimitsRecentHits(t *testing.T) {
	scanner := NewScanner(10, nil)
	scanner.Scan("<script>alert(1)</script>")
	scanner.Scan("union select username from users")

	stats := scanner.GetStats(1)
	if stats.TotalHits != 2 || len(stats.RecentHits) != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if stats.RecentHits[0].Pattern != "SQL Injection" {
		t.Fatalf("expected newest hit, got %#v", stats.RecentHits[0])
	}
}
