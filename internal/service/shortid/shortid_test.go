package shortid

import "testing"

func TestGenerateLength(t *testing.T) {
	value, err := Generate(12)
	if err != nil {
		t.Fatalf("generate short id: %v", err)
	}

	if len(value) != 12 {
		t.Fatalf("expected length 12, got %d", len(value))
	}
}

func TestGenerateDefaultLength(t *testing.T) {
	value, err := Generate(0)
	if err != nil {
		t.Fatalf("generate short id: %v", err)
	}

	if len(value) != 10 {
		t.Fatalf("expected default length 10, got %d", len(value))
	}
}
