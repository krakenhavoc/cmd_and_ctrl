package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// x_ceiling_test.go — the enumerator half of #2581: a printed "X can't
// be greater than <count>" caps the X every cast move announces and the
// open X range it carries (ADR 0122 §6.2), at the number CastSpell
// refuses above. Every move offered is accepted (#544).

const (
	oracleOpenTheWay   = "e07297c7-83bc-4162-bbb1-362bc737efe0"
	testSnowXCeilingID = "test-legal-x-ceiling-snow"
)

// withSnowXCeiling declares Winter's Chill's ceiling — the number of
// snow lands the caster controls — on a test card, through the
// catalog hook.
func withSnowXCeiling(t *testing.T) {
	t.Helper()
	prev := game.CatalogXCeiling
	game.CatalogXCeiling = func(key string) *game.XCeiling {
		if key == testSnowXCeilingID {
			return &game.XCeiling{
				Label: "the number of snow lands you control",
				Count: func(g *game.Game, caster uuid.UUID) int {
					return g.CountControlledMatchingForEffect(caster, game.PermanentQuery{
						Types: []string{"land"}, Supertypes: []string{"snow"},
					})
				},
			}
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogXCeiling = prev })
}

func snowCeilingSorcery(t *testing.T, g *game.Game, p *game.Player, snow, plain int) uuid.UUID {
	t.Helper()
	clearHand(p)
	id := handCard(p, game.Card{Name: "Test Chill", TypeLine: "Sorcery", ManaCost: "{X}{U}", OracleID: testSnowXCeilingID})
	for range snow {
		battlefieldCard(g, p, game.Card{Name: "Snow-Covered Island", TypeLine: "Basic Snow Land — Island"})
	}
	lands(g, p, "Island", "Island", plain)
	advanceTo(t, g, game.StepPrecombatMain)
	return id
}

// checkXRange asserts every cast move of `card` announces at most
// `ceiling` and carries no open X range above it.
func checkXRange(t *testing.T, casts []legal.Move, ceiling int) {
	t.Helper()
	for _, m := range casts {
		if x := xValueOf(t, m); x > ceiling {
			t.Errorf("move %q announces X=%d over the ceiling %d", m.Label, x, ceiling)
		}
		if m.Value != nil && m.Value.Max != nil && *m.Value.Max > ceiling {
			t.Errorf("move %q opens X up to %d over the ceiling %d", m.Label, *m.Value.Max, ceiling)
		}
	}
}

func TestXCeilingOfZeroOffersOnlyXZero(t *testing.T) {
	withSnowXCeiling(t)
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	card := snowCeilingSorcery(t, g, active, 0, 6)

	moves := legal.EnumerateFor(g, active.ID)
	casts := castMovesFor(moves, card)
	if len(casts) == 0 {
		t.Fatalf("no cast offered at X=0")
	}
	checkXRange(t, casts, 0)
	dispatchAll(t, g, active.ID, moves)
}

func TestXCeilingCapsTheAnnouncementBelowWhatTheManaPays(t *testing.T) {
	withSnowXCeiling(t)
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	// Eight lands pay for X=7; two of them are snow.
	card := snowCeilingSorcery(t, g, active, 2, 6)

	moves := legal.EnumerateFor(g, active.ID)
	casts := castMovesFor(moves, card)
	if len(casts) == 0 {
		t.Fatalf("no cast offered")
	}
	checkXRange(t, casts, 2)
	if x := xValueOf(t, casts[0]); x != 2 {
		t.Errorf("offered X=%d, want the ceiling 2 (the largest legal X)", x)
	}
	dispatchAll(t, g, active.ID, moves)
}

func TestXCeilingAboveWhatTheManaPaysChangesNothing(t *testing.T) {
	withSnowXCeiling(t)
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	// Three snow lands pay for X=2 at most.
	card := snowCeilingSorcery(t, g, active, 3, 0)

	moves := legal.EnumerateFor(g, active.ID)
	casts := castMovesFor(moves, card)
	if len(casts) == 0 {
		t.Fatalf("no cast offered")
	}
	if x := xValueOf(t, casts[0]); x != 2 {
		t.Errorf("offered X=%d, want 2 (what three lands pay for)", x)
	}
	dispatchAll(t, g, active.ID, moves)
}

// Open the Way, from the catalog: X can't be greater than the number
// of players in the game, which a concession lowers.
func TestOpenTheWayIsOfferedAtMostThePlayersInTheGame(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	card := handCard(active, game.Card{Name: "Open the Way", TypeLine: "Sorcery", ManaCost: "{X}{G}{G}", OracleID: oracleOpenTheWay})
	lands(g, active, "Forest", "Forest", 10)
	advanceTo(t, g, game.StepPrecombatMain)

	casts := castMovesFor(legal.EnumerateFor(g, active.ID), card)
	if len(casts) == 0 {
		t.Fatalf("no cast offered")
	}
	checkXRange(t, casts, 4)
	if x := xValueOf(t, casts[0]); x != 4 {
		t.Errorf("offered X=%d at four players, want 4", x)
	}

	if err := g.Concede(g.Seats[(g.Turn.ActiveSeat+1)%4].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	moves := legal.EnumerateFor(g, active.ID)
	casts = castMovesFor(moves, card)
	checkXRange(t, casts, 3)
	if len(casts) == 0 || xValueOf(t, casts[0]) != 3 {
		t.Errorf("after a concession: want X=3 offered, got %v", labels(casts))
	}
	dispatchAll(t, g, active.ID, moves)
}
