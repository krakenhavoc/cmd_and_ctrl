package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

// A draft seat's session (RoleDrafter) is handed to page script, so it
// must open nothing but its draft: Middleware refuses it on every route
// that does not name the role, including the "any session" ones.
func TestMiddlewareRefusesADraftSeatUnlessTheRouteNamesIt(t *testing.T) {
	a := NewMemoryAuthenticator()
	issue := func(p Principal) string {
		tok, _, err := a.Issue(context.Background(), p, time.Hour)
		if err != nil {
			t.Fatalf("Issue: %v", err)
		}
		return tok
	}
	drafter := issue(Principal{Role: RoleDrafter, DraftID: uuid.New(), PlayerID: uuid.New(), UserID: uuid.New()})
	player := issue(Principal{Role: RolePlayer, GameID: uuid.New(), PlayerID: uuid.New()})

	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	status := func(h http.Handler, tok string) int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	anySession := Middleware(a)(ok)
	if got := status(anySession, player); got != http.StatusNoContent {
		t.Fatalf("a player on an any-session route: %d", got)
	}
	if got := status(anySession, drafter); got != http.StatusForbidden {
		t.Fatalf("a draft seat on an any-session route: %d, want 403", got)
	}
	if got := status(Middleware(a, RolePlayer)(ok), drafter); got != http.StatusForbidden {
		t.Fatalf("a draft seat on a player route: %d, want 403", got)
	}
	if got := status(Middleware(a, RoleDrafter)(ok), drafter); got != http.StatusNoContent {
		t.Fatalf("a draft seat on a route that names it: %d", got)
	}
	if got := status(Middleware(a, RoleDrafter)(ok), player); got != http.StatusForbidden {
		t.Fatalf("a player on a draft-seat route: %d, want 403", got)
	}
}

// The draft a seat belongs to survives the signed token.
func TestHMACCarriesTheDraft(t *testing.T) {
	a := newTestHMAC(t, testKey)
	want := Principal{Role: RoleDrafter, DraftID: uuid.New(), PlayerID: uuid.New(), Name: "Sam"}
	tok, _, err := a.Issue(context.Background(), want, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	got, err := a.Validate(context.Background(), tok)
	if err != nil {
		t.Fatal(err)
	}
	if got.Role != RoleDrafter || got.DraftID != want.DraftID || got.PlayerID != want.PlayerID || got.GameID != uuid.Nil {
		t.Fatalf("round trip: %+v", got)
	}
}
