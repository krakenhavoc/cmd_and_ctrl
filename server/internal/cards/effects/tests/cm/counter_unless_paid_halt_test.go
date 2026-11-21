package testscm

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestAWardedPermanentCannotBeKilledForFreeByAPassingThirdSeat is
// #951's headline, card-side. Ward {3}, a Doom Blade from one
// opponent, and a THIRD seat passing before the payer has answered.
func TestAWardedPermanentCannotBeKilledForFreeByAPassingThirdSeat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := effectstest.PushCatalogPermanent(g, me.ID, "Rimeshield Frost Giant",
		"Creature — Giant Warrior", effectstest.RimeshieldOracle, false)

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := effectstest.CastAtWardedCreature(t, g, opp, giant)
	effectstest.PassUntilTaxed(t, g, opp.ID)

	// The move the bug was: somebody else passes.
	if err := effectstest.TheThirdSeatPasses(t, g); !errors.Is(err, game.ErrChoicePending) {
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
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sliver := effectstest.PushCatalogPermanent(g, me.ID, "Diffusion Sliver",
		"Creature — Sliver", effectstest.B41DiffusionSliverOracle, false)

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := effectstest.CastAtWardedCreature(t, g, opp, sliver)
	effectstest.PassUntilTaxed(t, g, opp.ID)

	if err := effectstest.TheThirdSeatPasses(t, g); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("a pass while the Sliver tax was open = %v, want ErrChoicePending", err)
	}
	if !g.Stack.Contains(blade) {
		t.Fatal("the removal resolved while the Sliver tax was unanswered")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sliver) {
		t.Error("a declined Diffusion Sliver tax must counter the spell")
	}
}

// The counterspell half of the same rule. Daze is "counter target
// spell unless its controller pays {1}", and it was wrong in exactly
// the way ward was: the spell it was answering could resolve while the
// {1} went unanswered.
func TestDazeHaltsTheStackWhileItsTaxIsOpen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)

	victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	daze := effectstest.HandCardFull(them, "Daze", "Instant", "{1}{U}", effectstest.DazeOracle, []string{"U"})
	island := effectstest.FreeSpellPermanent(g, them.ID, "Island", "Basic Land — Island")

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
	effectstest.PassUntilTaxed(t, g, me.ID)

	if err := effectstest.TheThirdSeatPasses(t, g); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("a pass while the Daze tax was open = %v, want ErrChoicePending", err)
	}
	if !g.Stack.Contains(victim) {
		t.Fatal("the Dazed spell resolved while the {1} was unanswered")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if g.Stack.Contains(victim) {
		t.Error("a declined Daze must counter the spell")
	}
}
