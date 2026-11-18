package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// urabrask_heretic_praetor_test.go — #5875: Urabrask, Heretic Praetor's
// own-upkeep impulse exile, and the one-shot draw replacement it sets up
// on each opponent (game/next_draw_exile_top.go), end to end through the
// turn engine.

func urabraskOnTable(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, owner, "Urabrask, Heretic Praetor", "Legendary Creature — Phyrexian Praetor",
		urabraskHereticPraetorOracleID, false)
}

// urabraskRecords counts the live next-draw replacements on `player`.
func urabraskRecords(g *game.Game, player uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, e := range g.ScopedEffects {
			for _, m := range e.Mods {
				if m.Kind == game.ModNextDrawExileTopPlay && m.Player == player {
					n++
				}
			}
		}
	})
	return n
}

// resolveUrabraskTrigger resolves the trigger Urabrask put on the stack at
// the current upkeep.
func resolveUrabraskTrigger(t *testing.T, g *game.Game, urabrask uuid.UUID) {
	t.Helper()
	if triggerOnStack(g, urabrask) == nil {
		t.Fatal("Urabrask did not trigger at this upkeep")
	}
	passPriorityAroundTable(t, g)
	if triggerOnStack(g, urabrask) != nil {
		t.Fatal("Urabrask's trigger did not resolve")
	}
}

func TestUrabraskHereticPraetorIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(urabraskHereticPraetorOracleID)
	if !ok {
		t.Fatal("Urabrask, Heretic Praetor is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %s, caveats %v", spec.Completeness, spec.Caveats)
	}
	if len(spec.Triggered) != 2 {
		t.Errorf("%d triggers, want 2", len(spec.Triggered))
	}
}

// The opponent's draw-step draw is replaced: the top card goes to exile,
// not to their hand, nothing is drawn, and the opponent (not Urabrask's
// controller) may play it. A second draw that turn is ordinary, and the
// permission is gone the next turn.
func TestUrabraskReplacesTheOpponentsFirstDrawOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	urabrask := urabraskOnTable(g, me.ID)

	advanceToUpkeepOf(t, g, 1)
	resolveUrabraskTrigger(t, g, urabrask)
	if got := urabraskRecords(g, opp.ID); got != 1 {
		t.Fatalf("%d replacements on the opponent after the trigger, want 1", got)
	}
	if got := urabraskRecords(g, me.ID); got != 0 {
		t.Fatalf("Urabrask's controller got %d replacements, want 0", got)
	}

	top := pushLibraryTopForTest(opp, "Urabrask Fodder")
	handBefore, libBefore := opp.Hand.Size(), opp.Library.Size()
	advanceToStepOf(t, g, 1, game.StepPrecombatMain)

	if !g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled in place of the draw")
	}
	if opp.Hand.Size() != handBefore || opp.Library.Size() != libBefore-1 {
		t.Errorf("hand %d (was %d), library %d (was %d): want the hand unchanged and one card gone from the library",
			opp.Hand.Size(), handBefore, opp.Library.Size(), libBefore)
	}
	if len(g.DrawnThisTurn[opp.ID]) != 0 {
		t.Error("a replaced draw drew a card")
	}
	if permissionFor(g, opp.ID, top) == nil {
		t.Error("the opponent may not play the exiled card")
	}
	if permissionFor(g, me.ID, top) != nil {
		t.Error("Urabrask's controller was given the opponent's exiled card")
	}
	if got := urabraskRecords(g, opp.ID); got != 0 {
		t.Errorf("the replacement was not spent: %d left", got)
	}

	// A second draw this turn is an ordinary draw.
	second := pushLibraryTopForTest(opp, "Second Draw")
	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatalf("second draw: %v", err)
	}
	if !opp.Hand.Contains(second) || g.Exile.Contains(second) {
		t.Error("the second draw this turn was replaced too")
	}

	// "This turn": the permission is gone on the next turn.
	advanceToUpkeepOf(t, g, 2)
	if permissionFor(g, opp.ID, top) != nil {
		t.Error("the opponent may still play the exiled card on a later turn")
	}
}

// A "draw two" is two draws (CR 121.2): the first is replaced, the
// second is drawn after the exile (CR 121.6b).
func TestUrabraskReplacesOnlyTheFirstOfADrawTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	urabrask := urabraskOnTable(g, me.ID)
	advanceToUpkeepOf(t, g, 1)
	resolveUrabraskTrigger(t, g, urabrask)

	second := pushLibraryTopForTest(opp, "Drawn Second")
	first := pushLibraryTopForTest(opp, "Exiled First")
	handBefore := opp.Hand.Size()
	var err error
	g.WithWriteLock(func() { err = g.DrawNForEffect(opp.ID, 2) })
	if err != nil {
		t.Fatalf("draw two: %v", err)
	}
	if !g.Exile.Contains(first) || permissionFor(g, opp.ID, first) == nil {
		t.Error("the first draw did not exile the top card with a play permission")
	}
	if !opp.Hand.Contains(second) || opp.Hand.Size() != handBefore+1 {
		t.Error("the second draw of the instruction did not draw the next card")
	}
}

// A turn in which the opponent draws nothing (Necropotence skips their
// draw step) ends with the replacement unused, and their next draw on a
// later turn is ordinary.
func TestUrabraskReplacementExpiresUnusedAtEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	urabrask := urabraskOnTable(g, me.ID)
	_ = seedReplacementPermanent(g, necropotenceOracle, "Necropotence", opp.ID)

	advanceToUpkeepOf(t, g, 1)
	resolveUrabraskTrigger(t, g, urabrask)
	if got := urabraskRecords(g, opp.ID); got != 1 {
		t.Fatalf("%d replacements on the opponent, want 1", got)
	}
	advanceToStepOf(t, g, 1, game.StepPrecombatMain)
	if got := urabraskRecords(g, opp.ID); got != 1 {
		t.Fatalf("the skipped draw step spent the replacement (%d left)", got)
	}

	advanceToUpkeepOf(t, g, 2)
	if got := urabraskRecords(g, opp.ID); got != 0 {
		t.Fatalf("the replacement outlived its turn: %d left", got)
	}
	card := pushLibraryTopForTest(opp, "Later Draw")
	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatalf("draw: %v", err)
	}
	if !opp.Hand.Contains(card) || g.Exile.Contains(card) {
		t.Error("a draw on a later turn was replaced")
	}
}

// An empty library: the replacement still applies (CR 121.6a), so nothing
// is drawn and the opponent does not lose for drawing from it.
func TestUrabraskReplacedDrawFromAnEmptyLibraryLosesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	urabrask := urabraskOnTable(g, me.ID)
	advanceToUpkeepOf(t, g, 1)
	resolveUrabraskTrigger(t, g, urabrask)

	opp.Library.Cards = nil
	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatalf("draw: %v", err)
	}
	if opp.AttemptedEmptyDraw {
		t.Error("a replaced draw from an empty library counted as drawing from it")
	}
	if got := urabraskRecords(g, opp.ID); got != 0 {
		t.Errorf("the replacement was not spent: %d left", got)
	}
}

// Urabrask's controller's own upkeep: they exile their top card and may
// play it this turn; their own draw is not replaced.
func TestUrabraskOwnUpkeepExilesTopPlayableThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	urabrask := urabraskOnTable(g, me.ID)

	advanceToUpkeepOf(t, g, 1)
	top := pushLibraryTopForTest(me, "Impulse Card")
	resolveUrabraskTrigger(t, g, urabrask)
	if !g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled")
	}
	if permissionFor(g, me.ID, top) == nil {
		t.Error("Urabrask's controller may not play the exiled card")
	}
	if got := urabraskRecords(g, me.ID); got != 0 {
		t.Errorf("Urabrask set up a draw replacement on its own controller (%d)", got)
	}

	drawn := pushLibraryTopForTest(me, "Own Draw")
	advanceToStepOf(t, g, 1, game.StepPrecombatMain)
	if !me.Hand.Contains(drawn) {
		t.Error("the controller's own draw step was replaced")
	}
	advanceToUpkeepOf(t, g, 2)
	if permissionFor(g, me.ID, top) != nil {
		t.Error("the permission outlived the turn")
	}
}
