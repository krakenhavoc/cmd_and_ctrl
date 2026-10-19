package main

import (
	"net/http"
	"testing"
)

// The catalogue is public (ADR 0042; S70, #2985): with sign-up gated on
// a Discord role (ADR 0156), a session on /catalog would put Scryfall
// data behind a chat server. /cards/, the general image proxy, keeps
// its session.
func TestCatalogIsPublicAndCardsStayGated(t *testing.T) {
	srv, _ := publicStack(t)
	if got := do(t, srv, "GET", "/catalog", "").StatusCode; got != http.StatusOK {
		t.Errorf("GET /catalog signed out = %d, want 200", got)
	}
	// Signed out, an out-of-scope id is the handler's own 404, not
	// auth.Middleware's 401.
	if got := do(t, srv, "GET", "/catalog/image/00000000-0000-0000-0000-000000000000", "").StatusCode; got != http.StatusNotFound {
		t.Errorf("GET /catalog/image/{id} signed out = %d, want 404", got)
	}
	if got := do(t, srv, "GET", "/cards/00000000-0000-0000-0000-000000000000/printings", "").StatusCode; got != http.StatusNotFound && got != http.StatusServiceUnavailable {
		t.Errorf("GET printing metadata signed out = %d, want public 404 or 503", got)
	}
	if got := do(t, srv, "GET", "/cards/x", "").StatusCode; got != http.StatusUnauthorized {
		t.Errorf("GET /cards/x signed out = %d, want 401", got)
	}
}
