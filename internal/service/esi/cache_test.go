package esi

import "testing"

func TestCharacterIDKeyNormalizesName(t *testing.T) {
	left := CharacterIDKey(" The Pilot ")
	right := CharacterIDKey("the pilot")
	if left != right {
		t.Fatalf("expected normalized names to produce same key, got %q and %q", left, right)
	}
}

func TestESIEntityKeys(t *testing.T) {
	if got, want := AffiliationKey(42), "esi:character:affiliation:42"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
	if got, want := EntityNameKey(99), "esi:entity:name:99"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
