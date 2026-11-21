package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestDoublecastCopiesTheNextInstantYouCast: Doublecast resolves with
// nothing to show for it, and the NEXT instant is copied — six damage
// off one Bolt.
func TestDoublecastCopiesTheNextInstantYouCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID
	before := effectstest.LifeOf(g, victim)

	effectstest.CastCatalogSpell(t, g, "Doublecast", "Sorcery", effectstest.DoublecastOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("Doublecast left %d delayed triggers, want 1", len(g.DelayedTriggers))
	}

	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me, victim, 1)

	if got := effectstest.LifeOf(g, victim); got != before-6 {
		t.Errorf("victim life = %d, want %d (3 from the copy, 3 from the Bolt)", got, before-6)
	}
	// The copy is not a card (CR 707.10): Doublecast and the Bolt are
	// the only two things in the graveyard.
	if got := effectstest.GraveyardSize(g, me); got != 2 {
		t.Errorf("graveyard = %d cards, want 2 — a copy is not a card", got)
	}
}

// TestDoublecastDoesNotCopyItself: "when you NEXT cast" cannot mean
// the cast that created it. The trigger is made while Doublecast
// resolves, and its own EventCast was emitted before that resolution
// began — so there is nothing to exclude and nothing that remembers
// to.
func TestDoublecastDoesNotCopyItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID

	effectstest.CastCatalogSpell(t, g, "Doublecast", "Sorcery", effectstest.DoublecastOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.GraveyardSize(g, me); got != 1 {
		t.Errorf("graveyard = %d cards after Doublecast alone, want 1", got)
	}
	if len(g.DelayedTriggers) != 1 {
		t.Errorf("Doublecast left %d delayed triggers, want 1 still owed", len(g.DelayedTriggers))
	}
}

// TestDoublecastIgnoresACreatureSpell: "an instant or sorcery spell"
// is the condition, and a cast that does not match leaves the trigger
// owed for the one that does.
func TestDoublecastIgnoresACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID

	effectstest.CastCatalogSpell(t, g, "Doublecast", "Sorcery", effectstest.DoublecastOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.CastCatalogSpell(t, g, "Sol Ring", "Artifact", effectstest.SolRingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("an artifact spell consumed the trigger (%d left)", len(g.DelayedTriggers))
	}

	before := effectstest.LifeOf(g, victim)
	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me, victim, 1)
	if got := effectstest.LifeOf(g, victim); got != before-6 {
		t.Errorf("victim life = %d, want %d — the trigger was still owed", got, before-6)
	}
}

// TestDoublecastCopiesOnlyOneSpell is CR 603.7b on the card: the
// trigger fires once, so a second instant in the same turn is not
// copied.
func TestDoublecastCopiesOnlyOneSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID

	effectstest.CastCatalogSpell(t, g, "Doublecast", "Sorcery", effectstest.DoublecastOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	before := effectstest.LifeOf(g, victim)
	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me, victim, 1)
	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me, victim, 0)

	if got := effectstest.LifeOf(g, victim); got != before-9 {
		t.Errorf("victim life = %d, want %d (3+3 then 3, not 3+3 then 3+3)", got, before-9)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d delayed triggers left owed, want 0", len(g.DelayedTriggers))
	}
}

// TestTwoGalvanicIterationsCopyTheSameSpellTwice: the card's whole
// storm-deck use, and the clearest statement that "once" is per
// trigger and not per event. Cast it, flash it back, then Bolt — two
// copies plus the original.
func TestTwoGalvanicIterationsCopyTheSameSpellTwice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID

	effectstest.CastCatalogSpell(t, g, "Galvanic Iteration", "Instant", effectstest.GalvanicIterationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// A second copy of the card rather than a flashback cast: the
	// flashback path is #411's and has its own tests; what this case
	// is about is two triggers owed at once.
	effectstest.CastCatalogSpell(t, g, "Galvanic Iteration", "Instant", effectstest.GalvanicIterationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 2 {
		t.Fatalf("%d delayed triggers owed, want 2", len(g.DelayedTriggers))
	}

	before := effectstest.LifeOf(g, victim)
	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me, victim, 2)

	if got := effectstest.LifeOf(g, victim); got != before-9 {
		t.Errorf("victim life = %d, want %d (two copies plus the Bolt)", got, before-9)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d delayed triggers left owed, want 0", len(g.DelayedTriggers))
	}
}

// TestDoublecastExpiresWithTheTurn: "this turn" (CR 514.2). A
// Doublecast that never found a spell is gone by the next turn, and
// the assertion fails with the cleanup sweep backed out.
func TestDoublecastExpiresWithTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := g.Seats[1].ID

	effectstest.CastCatalogSpell(t, g, "Doublecast", "Sorcery", effectstest.DoublecastOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("Doublecast left %d delayed triggers, want 1", len(g.DelayedTriggers))
	}

	startTurn := g.Turn.Seq
	for i := 0; i < 64 && g.Turn.Seq == startTurn; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("%d delayed triggers survived the turn, want 0", len(g.DelayedTriggers))
	}

	// And nothing is copied on the next turn's instant.
	next := g.Seats[g.Turn.ActiveSeat]
	before := effectstest.LifeOf(g, victim)
	boltID := uuid.New()
	next.Hand.PushTop(game.Card{
		InstanceID: boltID,
		Name:       "Lightning Bolt",
		TypeLine:   "Instant",
		OracleID:   effectstest.LightningBoltOracle,
		Owner:      next.ID,
		Controller: next.ID,
	})
	if err := g.CastSpell(next.ID, boltID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: victim}},
	}); err != nil {
		t.Fatalf("CastSpell Lightning Bolt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, victim); got != before-3 {
		t.Errorf("victim life = %d, want %d — an expired Doublecast copied a spell", got, before-3)
	}
}
