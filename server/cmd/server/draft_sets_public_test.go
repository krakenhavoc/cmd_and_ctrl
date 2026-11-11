package main

import (
	"net/http"
	"testing"
)

// The draftable-set list is public (ADR 0222 §4 and §9, #5437): it is
// card data, and a visitor can draft against bots without signing in, so
// a session on it would put card data behind the Discord sign-up gate.
// It is read-only.
func TestDraftSetsIsPublicAndReadOnly(t *testing.T) {
	srv, _ := publicStack(t)
	// publicStack has no dump: the answer is an empty list (draftpacks'
	// own tests read the body), not an error and not a 401.
	if got := do(t, srv, "GET", "/draft/sets", "").StatusCode; got != http.StatusOK {
		t.Fatalf("GET /draft/sets signed out = %d, want 200", got)
	}
	if got := do(t, srv, "POST", "/draft/sets", "").StatusCode; got != http.StatusMethodNotAllowed {
		t.Errorf("POST /draft/sets = %d, want 405", got)
	}
}
