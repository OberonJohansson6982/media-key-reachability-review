package main

import (
	"strings"
	"testing"
	"time"
)

func TestMediaReachability(t *testing.T) {
	required := map[string][]string{
		"asset ingestion":  {"asset:write"},
		"processing jobs":  {"asset:read", "job:write"},
		"creator delivery": {"asset:read", "delivery:read"},
	}
	tests := []struct {
		name   string
		scopes []string
		want   string
	}{
		{"ingest only", []string{"asset:write"}, "asset ingestion"},
		{"processing only", []string{"asset:read", "job:write"}, "processing jobs"},
		{"delivery only", []string{"asset:read", "delivery:read"}, "creator delivery"},
		{"missing shared read", []string{"job:write", "delivery:read"}, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rows := audit([]keyRecord{{ID: "k1", Name: "pipeline", Scopes: tc.scopes}}, required)
			got := strings.Join(rows[0].Reachable, ", ")
			if got != tc.want {
				t.Fatalf("reachable = %q, want %q", got, tc.want)
			}
			if !strings.Contains(reportMarkdown(time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC), rows), "2026-09-21") {
				t.Fatal("dated review missing")
			}
		})
	}
}
