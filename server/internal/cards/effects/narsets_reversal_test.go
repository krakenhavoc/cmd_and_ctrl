package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const narsetsReversalOracle = "d55f6c70-321f-4fb4-bd33-0850ae1a7c36"

// TestNarsetsReversalCopiesAndReturnsToHand: the original spell ends
// up back in its owner's hand (not countered into the graveyard), and
// a copy of it also resolves — proven by the copy's damage landing
// alongside the original's, once it is recast.
func TestNarsetsReversalCopiesAndReturnsToHand(t *testing.T) {
	g := newCatalogGame(t)
	caster, opponent := g.Seats[0], g.Seats[1]
	toMain(t, g)

	boltID := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: boltID, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: lightningBoltOracle, Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, boltID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opponent.ID}},
	}); err != nil {
		t.Fatalf("CastSpell Lightning Bolt: %v", err)
	}

	reversalID := uuid.New()
	opponent.Hand.PushTop(game.Card{
		InstanceID: reversalID, Name: "Narset's Reversal", TypeLine: "Instant",
		OracleID: narsetsReversalOracle, Owner: opponent.ID, Controller: opponent.ID,
	})
	if err := g.CastSpell(opponent.ID, reversalID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: boltID}},
	}); err != nil {
		t.Fatalf("CastSpell Narset's Reversal: %v", err)
	}

	lifeBefore := opponent.Life
	passPriorityAroundTable(t, g)

	// CR 707.10c: the copy carries an optional re-target prompt.
	// Re-pick the same player to keep the printed outcome — declining
	// (or re-picking) both mean "the copy keeps its targets".
	if p := latestPickTarget(g, opponent.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, opponent.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opponent.ID}); err != nil {
			t.Fatalf("ResolvePickTarget (re-target the copy): %v", err)
		}
		passPriorityAroundTable(t, g)
	}

	if g.Stack.Contains(boltID) {
		t.Error("the original Lightning Bolt should be off the stack")
	}
	if !caster.Hand.Contains(boltID) {
		t.Error("the original Lightning Bolt should be back in its owner's hand, not countered")
	}
	if caster.Graveyard.Contains(boltID) {
		t.Error("Narset's Reversal does not counter — the original must not land in a graveyard")
	}
	if opponent.Life != lifeBefore-3 {
		t.Errorf("opponent life %d -> %d, want -3 from the copy resolving", lifeBefore, opponent.Life)
	}
}

// TestNarsetsReversalIsNotACounter proves the same "no EventCounterSpell"
// shape Reprieve's test does — the original spell's departure from the
// stack must not be observable as a counter.
func TestNarsetsReversalIsNotACounter(t *testing.T) {
	g := newCatalogGame(t)
	caster, opponent := g.Seats[0], g.Seats[1]
	toMain(t, g)

	boltID := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: boltID, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: lightningBoltOracle, Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, boltID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opponent.ID}},
	}); err != nil {
		t.Fatalf("CastSpell Lightning Bolt: %v", err)
	}

	reversalID := uuid.New()
	opponent.Hand.PushTop(game.Card{
		InstanceID: reversalID, Name: "Narset's Reversal", TypeLine: "Instant",
		OracleID: narsetsReversalOracle, Owner: opponent.ID, Controller: opponent.ID,
	})
	eventsBefore := countEventsOf(g, game.EventCounterSpell, boltID)
	if err := g.CastSpell(opponent.ID, reversalID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: boltID}},
	}); err != nil {
		t.Fatalf("CastSpell Narset's Reversal: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := countEventsOf(g, game.EventCounterSpell, boltID); got != eventsBefore {
		t.Errorf("EventCounterSpell fired %d times, want 0 — Narset's Reversal is not a counter", got-eventsBefore)
	}
}
