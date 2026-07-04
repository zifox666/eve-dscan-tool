package dscan

import "testing"

func TestLocalCacheKey(t *testing.T) {
	got := LocalCacheKey("abc123")
	want := "dscan:local:abc123"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestShipCacheKey(t *testing.T) {
	tests := []struct {
		name    string
		lang    string
		shortID string
		want    string
	}{
		{name: "zh", lang: "zh", shortID: "abc123", want: "dscan:ship:zh:abc123"},
		{name: "en", lang: "en", shortID: "abc123", want: "dscan:ship:en:abc123"},
		{name: "fallback", lang: "ja", shortID: "abc123", want: "dscan:ship:zh:abc123"},
		{name: "empty", lang: "", shortID: "abc123", want: "dscan:ship:zh:abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShipCacheKey(tt.lang, tt.shortID)
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
