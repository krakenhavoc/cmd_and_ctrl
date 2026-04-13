package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func buildActiveGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := range 2 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i+1), uuid.Nil)}
		for j := range 10 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return g
}

func TestViewOfGameStructuralFields(t *testing.T) {
	g := buildActiveGame(t)
	v := ViewOfGame(g)

	if v.ID != g.ID.String() {
		t.Errorf("ID: got %q, want %q", v.ID, g.ID.String())
	}
	if v.State != "active" {
		t.Errorf("State: got %q, want %q", v.State, "active")
	}
	if len(v.Seats) != 2 {
		t.Errorf("Seats: got %d, want 2", len(v.Seats))
	}
	if v.Turn.Step != "untap" {
		t.Errorf("Turn.Step: got %q, want %q", v.Turn.Step, "untap")
	}
	if v.Battlefield.Kind != "battlefield" {
		t.Errorf("Battlefield.Kind: got %q", v.Battlefield.Kind)
	}
	if v.Battlefield.Owner != "" {
		t.Errorf("shared zone should have empty owner string, got %q", v.Battlefield.Owner)
	}
}

func TestViewOfGamePlayerContents(t *testing.T) {
	g := buildActiveGame(t)
	v := ViewOfGame(g)
	p := v.Seats[0]

	if p.Life != 40 {
		t.Errorf("Life: got %d, want 40", p.Life)
	}
	if p.Command.Count != 1 {
		t.Errorf("command zone count: got %d, want 1", p.Command.Count)
	}
	if p.Library.Count != 10 {
		t.Errorf("library count: got %d, want 10", p.Library.Count)
	}
	if p.Command.Owner == "" {
		t.Error("private zone should have non-empty owner")
	}
	if !p.Command.Cards[0].IsCommander {
		t.Error("commander view should have IsCommander set")
	}
}

func TestViewOfGameJSONRoundTrip(t *testing.T) {
	g := buildActiveGame(t)
	v := ViewOfGame(g)

	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back GameView
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.ID != v.ID {
		t.Errorf("id: got %q, want %q", back.ID, v.ID)
	}
	if back.Turn != v.Turn {
		t.Errorf("turn: got %+v, want %+v", back.Turn, v.Turn)
	}
	if len(back.Seats) != len(v.Seats) {
		t.Errorf("seats: got %d, want %d", len(back.Seats), len(v.Seats))
	}
}

func TestViewOfCardOmitsEmptyCounters(t *testing.T) {
	owner := uuid.New()
	c := game.NewCard("Sol Ring", owner)
	view := viewOfCard(c)

	raw, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// counters is omitempty — should not appear in the output at all.
	if contains := string(raw); containsKey(contains, "counters") {
		t.Errorf("counters field should be omitted when empty, got %s", contains)
	}
}

func containsKey(s, key string) bool {
	needle := `"` + key + `"`
	for i := 0; i+len(needle) <= len(s); i++ {
		if s[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

// buildViewWithCardsInHands creates a 2-seat active game, draws one
// card into each seat's hand, and returns the resulting GameView.
// Used by the FilterViewFor tests so that "hand.cards zeroed" is
// meaningful — in an empty game every hand is already empty.
func buildViewWithCardsInHands(t *testing.T) GameView {
	t.Helper()
	g := buildActiveGame(t)
	// Draw a card into each seat's hand directly on the game model.
	// The protocol package is allowed to reach into game.Card because
	// it's the wire boundary.
	for _, p := range g.Seats {
		if err := g.DrawCard(p.ID); err != nil {
			t.Fatalf("DrawCard: %v", err)
		}
	}
	return ViewOfGame(g)
}

func TestFilterViewForOwnSeatSeesHand(t *testing.T) {
	v := buildViewWithCardsInHands(t)
	own := v.Seats[0].ID

	filtered := FilterViewFor(v, own)

	if filtered.Seats[0].Hand.Count != 1 {
		t.Errorf("own hand count: got %d, want 1", filtered.Seats[0].Hand.Count)
	}
	if len(filtered.Seats[0].Hand.Cards) != 1 {
		t.Errorf("own hand.cards len: got %d, want 1", len(filtered.Seats[0].Hand.Cards))
	}
	// Own library cards also visible.
	if len(filtered.Seats[0].Library.Cards) == 0 {
		t.Error("own library.cards should be visible, got empty slice")
	}
}

func TestFilterViewForOpponentHandHidden(t *testing.T) {
	v := buildViewWithCardsInHands(t)
	own := v.Seats[0].ID

	filtered := FilterViewFor(v, own)

	// Opponent (seat 1): count preserved, cards zeroed.
	if filtered.Seats[1].Hand.Count != 1 {
		t.Errorf("opponent hand count: got %d, want 1 (count must survive filter)", filtered.Seats[1].Hand.Count)
	}
	if len(filtered.Seats[1].Hand.Cards) != 0 {
		t.Errorf("opponent hand.cards len: got %d, want 0 (must be hidden)", len(filtered.Seats[1].Hand.Cards))
	}
	// Opponent library contents hidden too (count preserved).
	if filtered.Seats[1].Library.Count == 0 {
		t.Error("opponent library count should be non-zero (preserved)")
	}
	if len(filtered.Seats[1].Library.Cards) != 0 {
		t.Errorf("opponent library.cards should be hidden, got %d", len(filtered.Seats[1].Library.Cards))
	}
	// Graveyard and Command zones remain visible.
	if filtered.Seats[1].Command.Count == 0 {
		t.Error("opponent command zone should be visible (count>0 at game start)")
	}
}

func TestFilterViewForSpectatorHidesAllHands(t *testing.T) {
	v := buildViewWithCardsInHands(t)

	// viewerID = "" is the spectator case: no seat matches, so every
	// seat is treated as an opponent.
	filtered := FilterViewFor(v, "")

	for i, seat := range filtered.Seats {
		if len(seat.Hand.Cards) != 0 {
			t.Errorf("seat %d hand.cards should be hidden from spectator, got %d", i, len(seat.Hand.Cards))
		}
	}
}

func TestFilterViewForDoesNotMutateInput(t *testing.T) {
	v := buildViewWithCardsInHands(t)
	own := v.Seats[0].ID

	origSeat1HandLen := len(v.Seats[1].Hand.Cards)
	_ = FilterViewFor(v, own)
	if got := len(v.Seats[1].Hand.Cards); got != origSeat1HandLen {
		t.Errorf("FilterViewFor mutated input: seat 1 hand.cards was %d, now %d", origSeat1HandLen, got)
	}
}

func TestFilterViewForZeroedHandMarshalsAsEmptyArray(t *testing.T) {
	// The hidden-hand zone must marshal to `"cards": []`, never
	// `"cards": null`. Client code treats `null` as "field missing"
	// and can trip on it.
	v := buildViewWithCardsInHands(t)
	own := v.Seats[0].ID
	filtered := FilterViewFor(v, own)

	raw, err := json.Marshal(filtered.Seats[1].Hand)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !bytes.Contains(raw, []byte(`"cards":[]`)) {
		t.Errorf("hidden hand must serialise with empty cards array, got: %s", raw)
	}
}

func TestActionPayloadRoundTrip(t *testing.T) {
	params, _ := json.Marshal(map[string]any{"delta": -3})
	a := ActionPayload{
		Type:   "change_life",
		Player: uuid.New().String(),
		Params: params,
	}
	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back ActionPayload
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.Type != a.Type || back.Player != a.Player {
		t.Errorf("action payload lost fields: %+v", back)
	}
}
