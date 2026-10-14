package lobby

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestAdminRatesASeatedLibraryDeck walks #3671: only an admin may rate
// a saved deck; a seat already holding it shows the rating at once, a
// reseat carries it, and a value off the scale or an unknown deck is
// refused.
func TestAdminRatesASeatedLibraryDeck(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	st := newDeckLibraryStack(t, idx)
	alice := mustLibraryUser(t, st.db, "Alice")

	source := "Commander:\n1 Test Commander\nMainboard:\n99 Plains\n"
	saved, err := st.library.Upsert(context.Background(), alice, "Alice's Deck", "text", source, []string{"Test Commander"}, 100)
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	meta, _ := st.lobby.Create("FNM")
	_, aliceSeat, _ := st.lobby.Join(meta.ID, meta.InviteToken, "Alice")
	aliceTok := signedInToken(t, st.auth, meta.ID, aliceSeat, alice)
	seatPath := "/games/" + meta.ID.String() + "/decks/" + saved.ID.String()
	powerPath := "/admin/decks/" + saved.ID.String() + "/power"

	resp := postJSON(t, st.srv, seatPath, aliceTok, struct{}{})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("seating the deck: %d", resp.StatusCode)
	}
	resp.Body.Close()
	if got := seatPower(t, st, meta.ID); got != 0 {
		t.Fatalf("an unrated deck's seat shows power %v; want 0", got)
	}

	// A player cannot rate, not even their own deck.
	resp = doJSON(t, st.srv, http.MethodPut, powerPath, aliceTok, map[string]any{"power_level": 5})
	if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("a player rating a deck: got %d, want 401 or 403", resp.StatusCode)
	}
	resp.Body.Close()

	adminResp := postJSON(t, st.srv, "/admin/login", "", adminLoginRequest{Token: "shared-admin-token"})
	var admin sessionResponse
	_ = json.NewDecoder(adminResp.Body).Decode(&admin)
	adminResp.Body.Close()

	resp = doJSON(t, st.srv, http.MethodPut, powerPath, admin.Token, map[string]any{"power_level": 3.5})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("admin rating: got %d, want 200", resp.StatusCode)
	}
	var out deckPowerResponse
	_ = json.NewDecoder(resp.Body).Decode(&out)
	resp.Body.Close()
	if out.DeckID != saved.ID.String() || out.Name != "Alice's Deck" || out.PowerLevel != 3.5 {
		t.Errorf("response = %+v", out)
	}
	if got := seatPower(t, st, meta.ID); got != 3.5 {
		t.Errorf("the seated deck shows power %v after rating; want 3.5", got)
	}

	// Reseating reads the rating from the deck.
	st.lobby.SetLibraryDeckPower(saved.ID.String(), 0)
	resp = postJSON(t, st.srv, seatPath, aliceTok, struct{}{})
	resp.Body.Close()
	if got := seatPower(t, st, meta.ID); got != 3.5 {
		t.Errorf("a reseated deck shows power %v; want 3.5", got)
	}

	for _, bad := range []any{0.5, 4.25, 6} {
		resp = doJSON(t, st.srv, http.MethodPut, powerPath, admin.Token, map[string]any{"power_level": bad})
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("power_level %v: got %d, want 400", bad, resp.StatusCode)
		}
		resp.Body.Close()
	}
	resp = doJSON(t, st.srv, http.MethodPut, "/admin/decks/"+uuid.New().String()+"/power", admin.Token, map[string]any{"power_level": 3})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("unknown deck: got %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()

	// null clears it, on the deck and on the seat.
	resp = doJSON(t, st.srv, http.MethodPut, powerPath, admin.Token, map[string]any{"power_level": nil})
	if resp.StatusCode != http.StatusOK {
		t.Errorf("clearing: got %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()
	if got := seatPower(t, st, meta.ID); got != 0 {
		t.Errorf("a cleared deck's seat shows power %v; want 0", got)
	}
}

// TestLoadGameReadsTheSeatedDecksPower: a restart brings a seat's
// rating back through the seats → decks join, and a seat with no
// library deck loads unrated.
func TestLoadGameReadsTheSeatedDecksPower(t *testing.T) {
	ctx := context.Background()
	st := newDeckLibraryStack(t, buildMinimalDeckIndex(t))
	alice := mustLibraryUser(t, st.db, "Alice")
	saved, err := st.library.Upsert(ctx, alice, "Alice's Deck", "text", "1 Test Commander", []string{"Test Commander"}, 100)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.library.SetPowerLevel(ctx, saved.ID, 4.5); err != nil {
		t.Fatal(err)
	}
	store := NewSQLStore(st.db)
	gameID := uuid.New()
	if err := store.CreateGame(ctx, GameRecord{ID: gameID, Name: "FNM", State: "lobby", CreatedAt: time.Now()}, nil); err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if err := store.ReplaceSeats(ctx, gameID, []SeatRecord{
		{Seat: 0, PlayerID: uuid.New(), GuestName: "Alice", DeckID: saved.ID.String(), DeckName: "Alice's Deck"},
		{Seat: 1, PlayerID: uuid.New(), GuestName: "Bob", DeckName: "Pasted"},
	}); err != nil {
		t.Fatalf("ReplaceSeats: %v", err)
	}
	_, seats, err := store.LoadGame(ctx, gameID)
	if err != nil {
		t.Fatalf("LoadGame: %v", err)
	}
	if len(seats) != 2 || seats[0].DeckPower != 4.5 || seats[1].DeckPower != 0 {
		t.Errorf("loaded seats = %+v; want power 4.5 then 0", seats)
	}
}

func seatPower(t *testing.T, st *deckLibraryStack, gameID uuid.UUID) float64 {
	t.Helper()
	g, err := st.lobby.Get(gameID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(g.Players) == 0 {
		t.Fatal("no seats")
	}
	return g.Players[0].DeckPower
}
