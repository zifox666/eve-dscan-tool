package esi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientCharacterIDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/universe/ids/" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.URL.Query().Get("datasource") != "tranquility" {
			t.Fatalf("missing datasource query")
		}

		var names []string
		if err := json.NewDecoder(r.Body).Decode(&names); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if len(names) != 1 || names[0] != "Pilot One" {
			t.Fatalf("unexpected names %#v", names)
		}

		_ = json.NewEncoder(w).Encode(universeIDsResponse{
			Characters: []universeIDCharacter{{ID: 123, Name: "Pilot One"}},
		})
	}))
	defer server.Close()

	client := NewClient(server.Client(), nil, ClientConfig{
		BaseURL:    server.URL,
		Datasource: "tranquility",
		Language:   "en",
		RetryDelay: time.Millisecond,
	})

	ids, err := client.CharacterIDs(t.Context(), []string{"Pilot One", "Pilot One", ""})
	if err != nil {
		t.Fatalf("character ids: %v", err)
	}
	if ids["Pilot One"] != 123 {
		t.Fatalf("expected id 123, got %d", ids["Pilot One"])
	}
}

func TestClientEntityNames(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/universe/names/" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		_ = json.NewEncoder(w).Encode([]EntityName{
			{ID: 1, Name: "Entity", Category: "character"},
		})
	}))
	defer server.Close()

	client := NewClient(server.Client(), nil, ClientConfig{
		BaseURL:    server.URL,
		Datasource: "tranquility",
		RetryDelay: time.Millisecond,
	})

	names, err := client.EntityNames(t.Context(), []int64{1, 1, 0})
	if err != nil {
		t.Fatalf("entity names: %v", err)
	}
	if names[1].Name != "Entity" {
		t.Fatalf("expected entity name, got %#v", names[1])
	}
}
