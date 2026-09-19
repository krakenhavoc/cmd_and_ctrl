package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_unless_paid_halt_test.go — #951 at the cards.
//
// The bug the issue reports is not visible in ward_test.go and that is
// the point of this file: those tests answer the prompt one pass at a
// time, which happens to dodge the interleaving. The bug needs a pass
// to arrive WHILE the prompt is outstanding — the ordinary case at a
// table with more than two seats, since any other player passing is
// enough. All four seats of newCatalogGame are used on purpose.

// theThirdSeatPasses is the move that used to kill a warded permanent
// for free: a player with no stake in the exchange passing priority
// while the payer has not answered.
func theThirdSeatPasses(t *testing.T, g *game.Game) error {
	t.Helper()
	return g.PassPriority()
}

// passUntilTaxed passes priority until `chooser` owes a pay-unless,
// stopping the moment one appears (or the gate refuses a pass, which
// is itself a prompt being open).
func passUntilTaxed(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	for i := 0; i < 8 && !hasPayUnlessFor(g, chooser); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority %d: %v", i, err)
		}
	}
	if !hasPayUnlessFor(g, chooser) {
		t.Fatalf("no pay-unless prompt ever reached %s", chooser)
	}
}

// TestAWardedPermanentCannotBeKilledForFreeByAPassingThirdSeat is
// #951's headline, card-side. Ward {3}, a Doom Blade from one
// opponent, and a THIRD seat passing before the payer has answered.
func TestAWardedPermanentCannotBeKilledForFreeByAPassingThirdSeat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := pushCatalogPermanent(g, me.ID, "Rimeshield Frost Giant",
		"Creature — Giant Warrior", rimeshieldOracle, false)

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := castAtWardedCreature(t, g, opp, giant)
	passUntilTaxed(t, g, opp.ID)

	// The move the bug was: somebody else passes.
	if err := theThirdSeatPasses(t, g); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("a pass while the ward tax was open = %v, want ErrChoicePending", err)
	}
	if _, err := g.AdvanceStep(); !errors.Is(err, game.ErrChoicePending) {
		t.Errorf("advance_step while the ward tax was open = %v, want ErrChoicePending", err)
	}
	if !g.Stack.Contains(blade) {
		t.Fatal("the Doom Blade resolved while the ward tax was unanswered — the ward was skipped")
	}
	if !g.Battlefield.Contains(giant) {
		t.Fatal("the warded creature died for free")
	}

	// And the tax is still settleable: declining counters the spell.
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(giant) {
		t.Error("a declined ward must counter the spell")
	}
	if !opp.Graveyard.Contains(blade) {
		t.Error("the countered spell should be in its owner's graveyard")
	}
}

// Diffusion Sliver is the card #951 was reproduced with: ward {2} for
// a whole board, riding the same pay-or-counter path.
func TestDiffusionSliverHaltsTheStackWhileItsTaxIsOpen(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sliver := pushCatalogPermanent(g, me.ID, "Diffusion Sliver",
		"Creature — Sliver", b41DiffusionSliverOracle, false)

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := castAtWardedCreature(t, g, opp, sliver)
	passUntilTaxed(t, g, opp.ID)

	if err := theThirdSeatPasses(t, g); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("a pass while the Sliver tax was open = %v, want ErrChoicePending", err)
	}
	if !g.Stack.Contains(blade) {
		t.Fatal("the removal resolved while the Sliver tax was unanswered")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sliver) {
		t.Error("a declined Diffusion Sliver tax must counter the spell")
	}
}

// The counterspell half of the same rule. Daze is "counter target
// spell unless its controller pays {1}", and it was wrong in exactly
// the way ward was: the spell it was answering could resolve while the
// {1} went unanswered.
func TestDazeHaltsTheStackWhileItsTaxIsOpen(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)

	victim := castCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	daze := handCardFull(them, "Daze", "Instant", "{1}{U}", dazeOracle, []string{"U"})
	island := freeSpellPermanent(g, them.ID, "Island", "Basic Land — Island")

	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the responder: %v", err)
	}
	if err := g.CastSpell(them.ID, daze, game.CastSpellParams{
		Strict:          true,
		AlternativeCost: "return",
		AltCostIDs:      []uuid.UUID{island},
		Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("casting Daze: %v", err)
	}
	passUntilTaxed(t, g, me.ID)

	if err := theThirdSeatPasses(t, g); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("a pass while the Daze tax was open = %v, want ErrChoicePending", err)
	}
	if !g.Stack.Contains(victim) {
		t.Fatal("the Dazed spell resolved while the {1} was unanswered")
	}
	answerPayUnless(t, g, me.ID, false)
	if g.Stack.Contains(victim) {
		t.Error("a declined Daze must counter the spell")
	}
}
