package legal_test

import (
	"encoding/json"
	"testing"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Thieving Skydiver, Verdeloth
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// kicker_x_test.go — the enumerator half of #3232: a kicker whose cost
// is {X} contributes its own X floor to the expansion that pays it, so
// the bot (which can only play what this list holds) is never offered a
// kicked Thieving Skydiver at X = 0, and an unkicked cast announces no
// X at all.

const (
	oracleThievingSkydiver = "f509ac02-f22e-4d1a-9fe1-afa96dd7806d"
	oracleVerdeloth        = "8222321f-b74d-4f74-9ee7-79455433345a"
)

func kickedOf(t *testing.T, m legal.Move) bool {
	t.Helper()
	var p struct {
		OptionalCosts []int `json:"optional_costs"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params %s: %v", string(m.Params), err)
	}
	return len(p.OptionalCosts) > 0
}

func TestKickedSkydiverIsNeverOfferedAtXZero(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	card := handCard(active, game.Card{Name: "Thieving Skydiver", TypeLine: "Creature — Merfolk Rogue",
		ManaCost: "{1}{U}", OracleID: oracleThievingSkydiver, Power: 2, Toughness: 1})
	lands(g, active, "Island", "Island", 5) // {1}{U} + X up to 3
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	var kicked, unkicked int
	for _, m := range castMovesFor(moves, card) {
		x := xValueOf(t, m)
		if kickedOf(t, m) {
			kicked++
			if x < 1 {
				t.Errorf("kicked move %q announces X=%d, want X >= 1", m.Label, x)
			}
			if m.Value != nil && m.Value.Min != nil && *m.Value.Min < 1 {
				t.Errorf("kicked move %q opens X from %d, want a floor of 1", m.Label, *m.Value.Min)
			}
			if x > 3 {
				t.Errorf("kicked move %q announces X=%d, more than the five lands pay", m.Label, x)
			}
			continue
		}
		unkicked++
		if x != 0 {
			t.Errorf("unkicked move %q announces X=%d, want none", m.Label, x)
		}
	}
	if kicked == 0 || unkicked == 0 {
		t.Fatalf("offered %d kicked and %d unkicked casts, want both", kicked, unkicked)
	}
	dispatchAll(t, g, active.ID, moves)
}

// With only the printed {1}{U} available, the kicked expansion cannot
// reach X = 1 and is not offered; the unkicked cast still is.
func TestKickedSkydiverIsNotOfferedWhenXOneIsUnaffordable(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	card := handCard(active, game.Card{Name: "Thieving Skydiver", TypeLine: "Creature — Merfolk Rogue",
		ManaCost: "{1}{U}", OracleID: oracleThievingSkydiver, Power: 2, Toughness: 1})
	lands(g, active, "Island", "Island", 2)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	casts := castMovesFor(moves, card)
	if len(casts) == 0 {
		t.Fatal("the unkicked cast was not offered")
	}
	for _, m := range casts {
		if kickedOf(t, m) {
			t.Errorf("kicked move %q offered with mana for X=0 only", m.Label)
		}
	}
	dispatchAll(t, g, active.ID, moves)
}

// A kicker with no floor (Verdeloth) still offers the kicked cast, at
// the largest affordable X.
func TestKickedVerdelothIsOfferedWithoutAFloor(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	card := handCard(active, game.Card{Name: "Verdeloth the Ancient", TypeLine: "Legendary Creature — Treefolk",
		ManaCost: "{4}{G}{G}", OracleID: oracleVerdeloth, Power: 4, Toughness: 7})
	lands(g, active, "Forest", "Forest", 8) // {4}{G}{G} + X = 2
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	best := -1
	for _, m := range castMovesFor(moves, card) {
		if kickedOf(t, m) {
			best = max(best, xValueOf(t, m))
		}
	}
	if best != 2 {
		t.Fatalf("largest kicked X offered = %d, want 2", best)
	}
	dispatchAll(t, g, active.ID, moves)
}
