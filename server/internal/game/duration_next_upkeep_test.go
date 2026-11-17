package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// duration_next_upkeep_test.go pins #5737: "until your next upkeep"
// (Erhnam Djinn) ends as the named player's next upkeep step BEGINS. Not
// as their next turn begins, a step earlier; not on a turn count, which
// cannot see a second upkeep in one turn or a skipped one; and, for a
// player who has left, as the turn holding that upkeep would have begun
// (CR 800.4m). Driven through the real step machine on a four-seat table.

func untilNextUpkeepLive(g *Game) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, e := range g.ScopedEffects {
			if e.Duration.Kind == UntilYourNextUpkeep {
				n++
			}
		}
	})
	return n
}

func registerUntilNextUpkeep(t *testing.T, g *Game, seat int) Duration {
	t.Helper()
	bear := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	var d Duration
	g.WithWriteLock(func() { d = g.UntilYourNextUpkeepDuration(g.Seats[seat].ID) })
	if d.Kind != UntilYourNextUpkeep || !d.Known() {
		t.Fatalf("duration = %+v, want a known until-your-next-upkeep window", d)
	}
	registerDurationMarker(g, bear, d, "until your next upkeep")
	return d
}

// Made in its player's own upkeep, the effect sits through every other
// seat's turn and through the beginning of its player's next turn, and
// ends as that turn's upkeep begins. The untap step is held open by a
// "you may choose not to untap" prompt so the test can stand in it: the
// turn has begun, so "until your next turn" would be over, and this is
// not.
func TestUntilYourNextUpkeepOutlastsTheUntapStepAndEndsAsTheUpkeepBegins(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	if g.Turn.Step != StepUpkeep || g.Turn.ActiveSeat != 0 {
		t.Fatalf("the table starts at %v of seat %d, want seat 0's upkeep", g.Turn.Step, g.Turn.ActiveSeat)
	}
	me := g.Seats[0]
	registerUntilNextUpkeep(t, g, 0)

	for seat := 1; seat <= 3; seat++ {
		advanceToStepOfSeat(t, g, seat, StepUpkeep)
		if untilNextUpkeepLive(g) != 1 {
			t.Fatalf("the effect ended during seat %d's turn", seat)
		}
	}

	const tick = "rust-tick-probe"
	pushTappedPermanent(g, me.ID, "Rust Tick", tick, "Artifact Creature", true)
	withCatalogUntapOptOuts(t, optOutSelf(tick))
	turnsBefore := me.TurnsBegun
	advanceToStepOfSeat(t, g, 0, StepUntap)
	prompt := openUntapChoice(t, g)
	if me.TurnsBegun != turnsBefore+1 {
		t.Fatalf("seat 0's turn has not begun (TurnsBegun %d, was %d)", me.TurnsBegun, turnsBefore)
	}
	if untilNextUpkeepLive(g) != 1 {
		t.Fatal("the effect ended as its player's turn began, before the upkeep")
	}
	if err := g.ResolveUntapChoice(prompt.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveUntapChoice: %v", err)
	}
	if g.Turn.Step != StepUpkeep || g.Turn.ActiveSeat != 0 {
		t.Fatalf("the cursor is at %v of seat %d, want seat 0's upkeep", g.Turn.Step, g.Turn.ActiveSeat)
	}
	if untilNextUpkeepLive(g) != 0 {
		t.Error("the effect outlived the beginning of its player's next upkeep")
	}
}

// The upkeep that ends it is the named player's: another player's upkeep
// ends nothing, and an effect made on someone else's turn waits for the
// named player's next one.
func TestUntilYourNextUpkeepWaitsForThatPlayersUpkeep(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)
	registerUntilNextUpkeep(t, g, 3)
	advanceToStepOfSeat(t, g, 2, StepUpkeep)
	if untilNextUpkeepLive(g) != 1 {
		t.Fatal("another player's upkeep ended the effect")
	}
	advanceToStepOfSeat(t, g, 3, StepUpkeep)
	if untilNextUpkeepLive(g) != 0 {
		t.Error("the effect outlived its player's upkeep")
	}
}

// A second upkeep in the same turn (CR 500.10, Obeka) is that player's
// next upkeep.
func TestUntilYourNextUpkeepEndsAtAnAdditionalUpkeepThisTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	registerUntilNextUpkeep(t, g, 0)
	advanceIntoStep(t, g, StepBeginCombat)
	addUpkeeps(t, g, me.ID, 1)
	advanceIntoStep(t, g, StepEndCombat)
	if untilNextUpkeepLive(g) != 1 {
		t.Fatal("the effect ended before the added upkeep")
	}
	advanceIntoStep(t, g, StepUpkeep)
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("the added upkeep is seat %d's, want seat 0's", g.Turn.ActiveSeat)
	}
	if untilNextUpkeepLive(g) != 0 {
		t.Error("the effect outlived the added upkeep")
	}
}

// A skipped turn has no upkeep (CR 614.10a), so the effect lasts on to
// the next upkeep that does begin.
func TestUntilYourNextUpkeepSurvivesASkippedTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	registerUntilNextUpkeep(t, g, 1)
	skipTurns(t, g, 1, uuid.Nil, 1)
	advanceToStepOfSeat(t, g, 2, StepUpkeep)
	if untilNextUpkeepLive(g) != 1 {
		t.Fatal("the effect ended in a turn that was skipped")
	}
	advanceToStepOfSeat(t, g, 1, StepUpkeep)
	if untilNextUpkeepLive(g) != 0 {
		t.Error("the effect outlived its player's next real upkeep")
	}
}

// CR 800.4m: a player who has left takes no upkeep, so the effect lasts
// until the turn holding that upkeep would have begun, and no longer.
func TestUntilYourNextUpkeepEndsWhenADepartedPlayersTurnWouldHaveBegun(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	leaver := g.Seats[2]
	registerUntilNextUpkeep(t, g, 2)
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if untilNextUpkeepLive(g) != 1 {
		t.Fatal("leaving the game ended the effect at once; CR 800.4m says it waits")
	}
	advanceOneTurn(t, g) // seat 1
	if untilNextUpkeepLive(g) != 1 {
		t.Fatal("the effect ended on seat 1's turn")
	}
	advanceOneTurn(t, g) // seat 2 is stepped over, to seat 3
	if g.Turn.ActiveSeat != 3 {
		t.Fatalf("active seat %d, want 3 after the departed seat", g.Turn.ActiveSeat)
	}
	if untilNextUpkeepLive(g) != 0 {
		t.Error("the departed player's turn would have begun and the effect is still live")
	}
}

// The effect is carried by a restore point and still ends at the right
// upkeep in the restored game.
func TestUntilYourNextUpkeepSurvivesARestore(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	registerUntilNextUpkeep(t, g, 0)
	advanceToStepOfSeat(t, g, 2, StepPrecombatMain)

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("not a restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	r, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if untilNextUpkeepLive(r) != 1 {
		t.Fatal("the restore dropped the effect")
	}
	advanceToStepOfSeat(t, r, 3, StepUpkeep)
	if untilNextUpkeepLive(r) != 1 {
		t.Fatal("the restored effect ended at another player's upkeep")
	}
	advanceToStepOfSeat(t, r, 0, StepUpkeep)
	if untilNextUpkeepLive(r) != 0 {
		t.Error("the restored effect outlived its player's next upkeep")
	}
}

func TestUntilYourNextUpkeepKeepsItsPersistedIntAndNamesItself(t *testing.T) {
	if !UntilYourNextUpkeep.Known() || UntilYourNextUpkeep != UntilNextEndStep+1 || (UntilYourNextUpkeep + 1).Known() {
		t.Error("UntilYourNextUpkeep is not the last known duration kind, right after UntilNextEndStep")
	}
	if int(UntilYourNextUpkeep) != 12 {
		t.Errorf("UntilYourNextUpkeep persists as %d, want 12", UntilYourNextUpkeep)
	}
	if got := UntilYourNextUpkeep.String(); got != "until your next upkeep" {
		t.Errorf("String() = %q", got)
	}
}
