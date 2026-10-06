package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// draw_instead_test.go — #2127 / #2168, the bot's half. A dredge offer
// is the ordinary yes/no "may" prompt and Underrealm Lich's pick is an
// ordinary choose_cards prompt, so the enumerator answers both with
// what it already has; this pins that it does, that declining is
// always legal, and that every answer it offers is accepted.

func TestDredgeOfferIsEnumeratedAndDecliningIsAlwaysLegal(t *testing.T) {
	g := newTable(t)
	p := g.Seats[1]
	g.WithWriteLock(func() {
		p.Graveyard.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Stinkweed Imp", TypeLine: "Creature — Imp",
			OracleID: "e005bc76-4985-4ec6-b9f6-cf6d0d9f5df4", Owner: p.ID, Controller: p.ID,
		})
	})
	if err := g.DrawCard(p.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	moves := legal.EnumerateFor(g, p.ID)
	if len(moves) != 2 {
		t.Fatalf("a dredge offer has two answers, enumerated %d: %v", len(moves), labels(moves))
	}
	var sawYes, sawNo bool
	for _, m := range moves {
		if m.Type != legal.TypeResolveChoice {
			t.Errorf("a seat owing a choice was offered %q", m.Label)
		}
		var params struct {
			Apply *bool `json:"apply"`
		}
		if err := json.Unmarshal(m.Params, &params); err != nil || params.Apply == nil {
			t.Fatalf("%q carries no yes/no answer: %s", m.Label, string(m.Params))
		}
		if *params.Apply {
			sawYes = true
		} else {
			sawNo = true
		}
	}
	if !sawYes || !sawNo {
		t.Errorf("both answers must be offered, so declining is always legal: yes=%v no=%v", sawYes, sawNo)
	}
	dispatchAll(t, g, p.ID, moves)
}

func TestUnderrealmLichPickIsEnumeratedAsOneOfThree(t *testing.T) {
	g := newTable(t)
	p := g.Seats[1]
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Underrealm Lich", TypeLine: "Creature — Zombie Elf Shaman",
			OracleID: "e1bce9c3-300c-4a9d-abe0-a1f02d3a1105", Power: 4, Toughness: 3,
			Owner: p.ID, Controller: p.ID,
		})
	})
	if err := g.DrawCard(p.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	moves := legal.EnumerateFor(g, p.ID)
	if len(moves) != 3 {
		t.Fatalf("the Lich's pick is exactly one of three cards, enumerated %d: %v", len(moves), labels(moves))
	}
	seen := map[uuid.UUID]bool{}
	for _, m := range moves {
		ids := pickedCardIDs(t, m)
		if len(ids) != 1 {
			t.Errorf("%q picks %d cards, want 1", m.Label, len(ids))
			continue
		}
		seen[ids[0]] = true
	}
	if len(seen) != 3 {
		t.Errorf("%d distinct cards offered, want 3", len(seen))
	}
	dispatchAll(t, g, p.ID, moves)
}
