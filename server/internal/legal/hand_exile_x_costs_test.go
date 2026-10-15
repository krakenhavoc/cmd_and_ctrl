package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// hand_exile_x_costs_test.go — #3175: the two hand-exile costs reach the
// move list, and every move they produce is one the engine accepts.

const (
	oracleShiningShoal        = "9937dae8-e639-46ad-849b-7e7be93bbbad"
	oracleMarchOfSwirlingMist = "debc69ea-372a-4720-838a-16856cd50b07"
)

type handExileCastParams struct {
	AlternativeCost  string   `json:"alternative_cost"`
	AltCostIDs       []string `json:"alt_cost_ids"`
	XValue           int      `json:"x_value"`
	ExileFromHandIDs []string `json:"exile_from_hand_ids"`
}

func decodeHandExile(t *testing.T, m legal.Move) handExileCastParams {
	t.Helper()
	var p handExileCastParams
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("decode %q: %v", m.Label, err)
	}
	return p
}

// Shining Shoal with no white mana: the only casts are pitches, one per
// distinct mana value of a white card in hand, each announcing that value
// as X ({X}{W} counts its X as 0 in hand, so it is mana value 1), and a
// blue card never pays.
func TestShiningShoalPitchIsOfferedOncePerManaValueWithThatX(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	shoal := handCard(active, game.Card{
		Name: "Shining Shoal", TypeLine: "Instant — Arcane", OracleID: oracleShiningShoal, ManaCost: "{X}{W}{W}",
	})
	mv := map[string]int{}
	for _, c := range []game.Card{
		{Name: "Savannah Lions", TypeLine: "Creature — Cat", ManaCost: "{W}"},
		{Name: "Serra Angel", TypeLine: "Creature — Angel", ManaCost: "{3}{W}{W}"},
		{Name: "Another Angel", TypeLine: "Creature — Angel", ManaCost: "{3}{W}{W}"},
		{Name: "Blue Drake", TypeLine: "Creature — Drake", ManaCost: "{1}{U}"},
		{Name: "White X Spell", TypeLine: "Sorcery", ManaCost: "{X}{W}"},
	} {
		id := handCard(active, c)
		mv[id.String()] = map[string]int{"{W}": 1, "{3}{W}{W}": 5, "{1}{U}": -1, "{X}{W}": 1}[c.ManaCost]
	}
	advanceTo(t, g, game.StepPrecombatMain)

	moves := castMovesFor(legal.EnumerateFor(g, active.ID), shoal)
	if len(moves) == 0 {
		t.Fatal("Shining Shoal is not offered by its pitch")
	}
	xs := map[int]bool{}
	for _, m := range moves {
		p := decodeHandExile(t, m)
		if p.AlternativeCost != "pitch" {
			t.Errorf("%q is not a pitch, with no mana to hard-cast", m.Label)
			continue
		}
		if len(p.AltCostIDs) != 1 {
			t.Fatalf("%q pitches %d cards, want 1", m.Label, len(p.AltCostIDs))
		}
		want := mv[p.AltCostIDs[0]]
		if want < 1 || p.XValue != want {
			t.Errorf("%q pitches a card of mana value %d at X = %d", m.Label, want, p.XValue)
		}
		xs[p.XValue] = true
	}
	if !xs[1] || !xs[5] || len(xs) != 2 {
		t.Errorf("X values offered = %v, want exactly 1 and 5", xs)
	}
	dispatchAll(t, g, active.ID, moves)
}

// March of Swirling Mist with one Island: unaffordable at X >= 1 without
// the discount, so every offered cast exiles blue cards, never the red
// one, and the engine accepts each.
func TestMarchOfSwirlingMistIsOfferedWithItsHandExileDiscount(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	march := handCard(active, game.Card{
		Name: "March of Swirling Mist", TypeLine: "Instant", OracleID: oracleMarchOfSwirlingMist, ManaCost: "{X}{U}",
	})
	blue := map[string]bool{}
	for range 2 {
		blue[handCard(active, game.Card{Name: "Blue Drake", TypeLine: "Creature — Drake", ManaCost: "{1}{U}"}).String()] = true
	}
	red := handCard(active, game.Card{Name: "Red Ogre", TypeLine: "Creature — Ogre", ManaCost: "{2}{R}"})
	lands(g, active, "Island", "Island", 1)
	for range 2 {
		battlefieldCard(g, opp, creature("Bear", "{1}{G}", 2, 2))
	}
	advanceTo(t, g, game.StepPrecombatMain)

	moves := castMovesFor(legal.EnumerateFor(g, active.ID), march)
	if len(moves) == 0 {
		t.Fatal("March of Swirling Mist is not offered with blue cards to exile")
	}
	for _, m := range moves {
		p := decodeHandExile(t, m)
		if len(p.ExileFromHandIDs) == 0 {
			t.Errorf("%q exiles nothing, but one Island cannot pay X >= 1 and {U}", m.Label)
		}
		for _, id := range p.ExileFromHandIDs {
			if id == red.String() || !blue[id] {
				t.Errorf("%q exiles %s, which is not a blue card in hand", m.Label, id)
			}
		}
		if p.XValue > 2*len(p.ExileFromHandIDs) {
			t.Errorf("%q announces X = %d with %d cards exiled and one Island", m.Label, p.XValue, len(p.ExileFromHandIDs))
		}
	}
	dispatchAll(t, g, active.ID, moves)
}

// A card without the cost is offered exactly as before: one payment, no
// exile list.
func TestNoHandExilePaymentsWithoutTheCost(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	shoal := handCard(active, game.Card{
		Name: "Shining Shoal", TypeLine: "Instant — Arcane", OracleID: oracleShiningShoal, ManaCost: "{X}{W}{W}",
	})
	handCard(active, game.Card{Name: "Blue Drake", TypeLine: "Creature — Drake", ManaCost: "{1}{U}"})
	lands(g, active, "Plains", "Plains", 3)
	advanceTo(t, g, game.StepPrecombatMain)
	moves := castMovesFor(legal.EnumerateFor(g, active.ID), shoal)
	if len(moves) == 0 {
		t.Fatal("Shining Shoal is not offered for its printed cost")
	}
	for _, m := range moves {
		if p := decodeHandExile(t, m); len(p.ExileFromHandIDs) > 0 {
			t.Errorf("%q names exile_from_hand_ids on a spell with no hand-exile cost", m.Label)
		}
	}
}
