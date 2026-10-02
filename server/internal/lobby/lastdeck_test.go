package lobby

// Tests for ADR 0110 section 5 item 5 (Delivery PR 7): the deck a
// signed-in person last seated is recorded and served by GET
// /me/last-deck, for the deck panel to preselect.

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decks"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
)

func (s *tableStack) lastDeck(t *testing.T, tok string) *users.LastDeck {
	t.Helper()
	resp := doGet(t, s.srv, "/me/last-deck", tok)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("GET /me/last-deck: %d %s", resp.StatusCode, raw)
	}
	var out lastDeckResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out.LastDeck
}

func TestLastDeckIsRecordedWhenASignedInPlayerSeatsADeck(t *testing.T) {
	s := newTableStack(t)
	alice, _ := s.signIn(t, "discord-99", "Alice")
	if got := s.lastDeck(t, alice); got != nil {
		t.Fatalf("before any deck: %+v", got)
	}

	table := s.mustCreate(t, s.admin(t), "decks")
	seat := s.join(t, alice, table.GameMeta, "")
	gamePath := "/games/" + table.ID.String() + "/decks"

	// A pasted list, saved to the library: the library deck.
	resp := postJSON(t, s.srv, gamePath, seat.Token,
		uploadDeckRequest{Format: "text", Source: testDeckSource, PlayerID: seat.PlayerID})
	var uploaded uploadDeckResponse
	_ = json.NewDecoder(resp.Body).Decode(&uploaded)
	resp.Body.Close()
	got := s.lastDeck(t, alice)
	if got == nil || got.Kind != users.LastDeckLibrary {
		t.Fatalf("after a paste: %+v", got)
	}
	libraryID := got.ID

	// A pre-built deck: the pre-built id.
	prebuilt := decks.All()[0].ID
	resp = postJSON(t, s.srv, gamePath, seat.Token,
		uploadDeckRequest{Deck: prebuilt, PlayerID: seat.PlayerID})
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("prebuilt: %d %s", resp.StatusCode, raw)
	}
	resp.Body.Close()
	if got := s.lastDeck(t, alice); got == nil || *got != (users.LastDeck{Kind: users.LastDeckPrebuilt, ID: prebuilt}) {
		t.Fatalf("after a pre-built deck: %+v", got)
	}

	// A library deck seated again: the library deck.
	if code := status(t, s.srv, "POST", gamePath+"/"+libraryID, seat.Token, nil); code != http.StatusOK {
		t.Fatalf("seat library deck: %d", code)
	}
	if got := s.lastDeck(t, alice); got == nil || *got != (users.LastDeck{Kind: users.LastDeckLibrary, ID: libraryID}) {
		t.Fatalf("after a library deck: %+v", got)
	}

	// The same seat reads it as the identity session does.
	if got := s.lastDeck(t, seat.Token); got == nil || got.ID != libraryID {
		t.Errorf("from the seat session: %+v", got)
	}
}

// TestLastDeckIsNotRecordedForSomeoneElse: an admin installing a deck
// on another person's seat does not record it as anyone's last deck.
func TestLastDeckIsNotRecordedForSomeoneElse(t *testing.T) {
	s := newTableStack(t)
	alice, _ := s.signIn(t, "discord-99", "Alice")
	table := s.mustCreate(t, s.admin(t), "decks")
	seat := s.join(t, alice, table.GameMeta, "")
	resp := postJSON(t, s.srv, "/games/"+table.ID.String()+"/decks", s.admin(t),
		uploadDeckRequest{Deck: decks.All()[0].ID, PlayerID: seat.PlayerID})
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("admin install: %d %s", resp.StatusCode, raw)
	}
	resp.Body.Close()
	if got := s.lastDeck(t, alice); got != nil {
		t.Errorf("an admin's install was recorded as Alice's: %+v", got)
	}
}

func TestGetLastDeckRefusesAGuest(t *testing.T) {
	s := newTableStack(t)
	table := s.mustCreate(t, s.admin(t), "t")
	guest := s.join(t, "", table.GameMeta, "Guest")
	for _, who := range []struct {
		name, token string
		want        int
	}{
		{"no credential", "", http.StatusUnauthorized},
		{"guest seat", guest.Token, http.StatusForbidden},
		{"admin token", s.admin(t), http.StatusForbidden},
	} {
		if got := status(t, s.srv, "GET", "/me/last-deck", who.token, nil); got != who.want {
			t.Errorf("%s: %d, want %d", who.name, got, who.want)
		}
	}
}
