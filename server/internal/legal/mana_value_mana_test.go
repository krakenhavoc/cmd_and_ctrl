package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// mana_value_mana_test.go — #4694, the enumerator half of Ashling,
// Rimebound's "Spend this mana only to cast spells with mana value 4 or
// greater" (CR 106.6, CR 202.3e). The engine half is
// game/mana_restriction_mana_value_test.go.

// With four such mana floating and nothing else, a four-drop is a move,
// a three-drop is not, and an X spell is offered at the X that makes
// its mana value 4 or more: an X search that stopped at the first X it
// could not pay would find none, since {X}{R} at X = 0 can't spend it.
func TestManaValueManaMakesOnlyBigEnoughSpellsMoves(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	three := handCard(active, game.Card{Name: "Three Drop", TypeLine: "Instant", ManaCost: "{2}{R}"})
	four := handCard(active, game.Card{Name: "Four Drop", TypeLine: "Instant", ManaCost: "{2}{R}{R}"})
	fireball := handCard(active, game.Card{Name: "Test Fireball", TypeLine: "Instant", ManaCost: "{X}{R}"})
	advanceTo(t, g, game.StepPrecombatMain)
	restricted := []string{game.ManaRestrictManaValueAtLeast(4)}
	for range 4 {
		active.ManaPool.AddMana(game.ManaToken{Color: "R", Restrictions: restricted})
	}

	moves := legal.EnumerateFor(g, active.ID)
	if len(castMovesOf(moves, three)) != 0 {
		t.Error("a three-drop is offered for mana spendable only on mana value 4 or greater")
	}
	if len(castMovesOf(moves, four)) == 0 {
		t.Error("a four-drop is not offered for the mana")
	}
	xs := castMovesOf(moves, fireball)
	if len(xs) == 0 {
		t.Fatal("{X}{R} is not offered: X = 3 makes it mana value 4")
	}
	for _, m := range xs {
		var p struct {
			XValue int `json:"x_value"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if p.XValue != 3 {
			t.Errorf("%q announces X = %d, want 3 (all four mana, mana value 4)", m.Label, p.XValue)
		}
	}
	dispatchAll(t, g, active.ID, castMovesOf(moves, four))
	dispatchAll(t, g, active.ID, xs)
}

// The same mana beside two Mountains: X = 1 is paid by the Mountains
// alone, X = 2 can't be (mana value 3), and X = 5 spends everything.
func TestManaValueManaJumpsTheXSearchPastTheThreshold(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	fireball := handCard(active, game.Card{Name: "Test Fireball", TypeLine: "Instant", ManaCost: "{X}{R}"})
	lands(g, active, "Mountain", "Mountain", 2)
	advanceTo(t, g, game.StepPrecombatMain)
	restricted := []string{game.ManaRestrictManaValueAtLeast(4)}
	for range 4 {
		active.ManaPool.AddMana(game.ManaToken{Color: "R", Restrictions: restricted})
	}
	xs := castMovesOf(legal.EnumerateFor(g, active.ID), fireball)
	if len(xs) == 0 {
		t.Fatal("{X}{R} is not offered")
	}
	for _, m := range xs {
		var p struct {
			XValue int `json:"x_value"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if p.XValue != 5 {
			t.Errorf("%q announces X = %d, want 5", m.Label, p.XValue)
		}
	}
	dispatchAll(t, g, active.ID, xs)
}
