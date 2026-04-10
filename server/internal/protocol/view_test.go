package protocol

import (
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
