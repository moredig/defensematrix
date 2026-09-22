package detection

import (
	"testing"
	"time"
)

func TestScannerScansAndCopiesHits(t *testing.T) {
	triggered := make(chan struct{}, 1)
	scanner := NewScanner(2, func() { triggered <- struct{}{} })

	scanner.Scan("union select username from users")
	hits := scanner.GetHits()
	if len(hits) != 1 || hits[0].Pattern != "SQL Injection" {
		t.Fatalf("unexpected hits: %#v", hits)
	}

	hits[0].Payload = "changed"
	if scanner.GetHits()[0].Payload == "changed" {
		t.Fatal("GetHits returned mutable scanner state")
	}

	scanner.Scan("nmap -sV 127.0.0.1")
	select {
	case <-triggered:
	case <-time.After(time.Second):
		t.Fatal("expected threshold callback")
	}
}

func TestScannerClampsInvalidThreshold(t *testing.T) {
	triggered := make(chan struct{}, 1)
	scanner := NewScanner(0, func() { triggered <- struct{}{} })
	scanner.Scan("nmap")
	select {
	case <-triggered:
	case <-time.After(time.Second):
		t.Fatal("expected zero threshold to clamp to one")
	}
}
