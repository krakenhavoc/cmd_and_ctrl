package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// next_cast_copy_test.go — the catalog half of #663: the two cards
// whose printed ability is a delayed trigger with an EVENT condition.
// The mechanism (fires once, expires with the turn, takes the
// harvester's dispatch, survives a clone) is pinned in
// game/delayed_on_event_test.go.

const (
	doublecastOracle        = "946675f5-8998-4f1b-934b-85ffe6e2f002"
	galvanicIterationOracle = "c51ed6e3-813b-49c4-b1de-7f92a77f8f6e"
)

// castBoltAtAndKeepTheCopysTarget casts a Lightning Bolt at `victim`,
// then answers the CR 707.10c "you may choose new targets" prompt for
// every copy the cast produced by keeping the original target, and
// settles the stack.
//
// Keeping the target is the interesting answer here: the point of
// these cards is the second three damage, not the re-aim, and
// Reverberate's test already pins the re-aim.
func castBoltAtAndKeepTheCopysTarget(t *testing.T, g *game.Game, caster, victim uuid.UUID, copies int) {
	t.Helper()
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})

	answered := 0
	for i := 0; i < 48; i++ {
		// Two owed triggers go on the stack at one boundary, which is
		// an ordinary CR 603.3b ordering prompt — ordinary because by
		// then these ARE ordinary triggers.
		if p := triggerOrderPromptFor(g, caster); p != nil {
			order := append([]uuid.UUID(nil), p.TriggerOrderIDs...)
			if err := g.ResolveTriggerOrder(p.ID, caster, order); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if p := latestPickTarget(g, caster); p != nil {
			if err := g.ResolvePickTarget(p.ID, caster,
				game.TargetRef{Kind: game.TargetPlayer, ID: victim}); err != nil {
				t.Fatalf("ResolvePickTarget for copy %d: %v", answered+1, err)
			}
			answered++
			continue
		}
		if stackFullyEmpty(g) {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if answered != copies {
		t.Fatalf("answered %d copy-target prompts, want %d", answered, copies)
	}
	passPriorityAroundTable(t, g)
}

// triggerOrderPromptFor is the outstanding CR 603.3b ordering prompt
// addressed to `chooser`, or nil.
func triggerOrderPromptFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerOrder && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// TestDoublecastCopiesTheNextInstantYouCast: Doublecast resolves with
// nothing to show for it, and the NEXT instant is copied — six damage
// off one Bolt.
func TestDoublecastCopiesTheNextInstantYouCast(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID
	before := lifeOf(g, victim)

	castCatalogSpell(t, g, "Doublecast", "Sorcery", doublecastOracle, nil)
	passPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("Doublecast left %d delayed triggers, want 1", len(g.DelayedTriggers))
	}

	castBoltAtAndKeepTheCopysTarget(t, g, me, victim, 1)

	if got := lifeOf(g, victim); got != before-6 {
		t.Errorf("victim life = %d, want %d (3 from the copy, 3 from the Bolt)", got, before-6)
	}
	// The copy is not a card (CR 707.10): Doublecast and the Bolt are
	// the only two things in the graveyard.
	if got := graveyardSize(g, me); got != 2 {
		t.Errorf("graveyard = %d cards, want 2 — a copy is not a card", got)
	}
}

// TestDoublecastDoesNotCopyItself: "when you NEXT cast" cannot mean
// the cast that created it. The trigger is made while Doublecast
// resolves, and its own EventCast was emitted before that resolution
// began — so there is nothing to exclude and nothing that remembers
// to.
func TestDoublecastDoesNotCopyItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID

	castCatalogSpell(t, g, "Doublecast", "Sorcery", doublecastOracle, nil)
	passPriorityAroundTable(t, g)

	if got := graveyardSize(g, me); got != 1 {
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
	g := newCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID

	castCatalogSpell(t, g, "Doublecast", "Sorcery", doublecastOracle, nil)
	passPriorityAroundTable(t, g)

	castCatalogSpell(t, g, "Sol Ring", "Artifact", solRingOracle, nil)
	passPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("an artifact spell consumed the trigger (%d left)", len(g.DelayedTriggers))
	}

	before := lifeOf(g, victim)
	castBoltAtAndKeepTheCopysTarget(t, g, me, victim, 1)
	if got := lifeOf(g, victim); got != before-6 {
		t.Errorf("victim life = %d, want %d — the trigger was still owed", got, before-6)
	}
}

// TestDoublecastCopiesOnlyOneSpell is CR 603.7b on the card: the
// trigger fires once, so a second instant in the same turn is not
// copied.
func TestDoublecastCopiesOnlyOneSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID

	castCatalogSpell(t, g, "Doublecast", "Sorcery", doublecastOracle, nil)
	passPriorityAroundTable(t, g)

	before := lifeOf(g, victim)
	castBoltAtAndKeepTheCopysTarget(t, g, me, victim, 1)
	castBoltAtAndKeepTheCopysTarget(t, g, me, victim, 0)

	if got := lifeOf(g, victim); got != before-9 {
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
	g := newCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID

	castCatalogSpell(t, g, "Galvanic Iteration", "Instant", galvanicIterationOracle, nil)
	passPriorityAroundTable(t, g)
	// A second copy of the card rather than a flashback cast: the
	// flashback path is #411's and has its own tests; what this case
	// is about is two triggers owed at once.
	castCatalogSpell(t, g, "Galvanic Iteration", "Instant", galvanicIterationOracle, nil)
	passPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 2 {
		t.Fatalf("%d delayed triggers owed, want 2", len(g.DelayedTriggers))
	}

	before := lifeOf(g, victim)
	castBoltAtAndKeepTheCopysTarget(t, g, me, victim, 2)

	if got := lifeOf(g, victim); got != before-9 {
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
	g := newCatalogGame(t)
	victim := g.Seats[1].ID

	castCatalogSpell(t, g, "Doublecast", "Sorcery", doublecastOracle, nil)
	passPriorityAroundTable(t, g)
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
	before := lifeOf(g, victim)
	boltID := uuid.New()
	next.Hand.PushTop(game.Card{
		InstanceID: boltID,
		Name:       "Lightning Bolt",
		TypeLine:   "Instant",
		OracleID:   lightningBoltOracle,
		Owner:      next.ID,
		Controller: next.ID,
	})
	if err := g.CastSpell(next.ID, boltID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: victim}},
	}); err != nil {
		t.Fatalf("CastSpell Lightning Bolt: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, victim); got != before-3 {
		t.Errorf("victim life = %d, want %d — an expired Doublecast copied a spell", got, before-3)
	}
}
