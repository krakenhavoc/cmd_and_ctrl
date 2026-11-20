package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pay_unless_resolving_spell_test.go — #5793: a PayUnless asked by a
// RESOLVING SPELL holds the table while "that player" decides.
//
// The spell below is Rhystic Deluge's and Withdraw's clause on a sorcery:
// "Return target creature to its owner's hand unless its controller pays
// {1}. Draw a card." The draw is a later instruction (CR 608.2c), so it
// shows whether the rest of the resolution waits for the answer.
//
// The report was reproduced on f2353f118: through the plain
// effects.PayUnless the spell finished before the payer answered (the draw
// ran, the sorcery went to the graveyard, any seat could pass and the step
// could end), and the bounce landed whenever the payer answered. The owner
// decided (2026-11-19, option 1) that the engine derives the halt: a
// pay-unless a resolution asks is owed in the step it is asked in (CR
// 118.12: the payment is a cost paid as the spell resolves; CR 608.2g;
// CR 117.3b). These tests are the same scenarios with the assertions
// turned round.

const payUnlessSpellOracle = "test-5793-pay-unless-spell"

func init() {
	Register(Spec{
		OracleID: payUnlessSpellOracle,
		Name:     "Fixture Rhystic Return",
		Targets:  TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(askTheTargetsControllerForOne),
			DrawCards{N: 1},
		),
	})
}

// askTheTargetsControllerForOne asks the target creature's controller for
// {1} through the plain PayUnless; the decline is Crystal Shard's
// registered bounce, so no test-only key is added to the ledger. Nothing
// here says "hold the table": the engine works that out.
func askTheTargetsControllerForOne(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		question := "pay {1}, or " + c.Name + " returns to its owner's hand?"
		return PayUnless{Chooser: c.Controller, Cost: "{1}", Question: question, Else: crystalShardBounce}.Apply(ctx)
	}
	return nil
}

// castRhysticReturn casts the fixture spell from seat 0 at a creature seat
// 1 controls, resolves it, and returns the table with the payer's question
// open: the spell, the creature and seat 0's hand size just before the
// spell resolved.
func castRhysticReturn(t *testing.T) (g *game.Game, spell, bear uuid.UUID, handBefore int) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear = pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	spell = castCatalogSpell(t, g, "Fixture Rhystic Return", "Sorcery", payUnlessSpellOracle, cardRefs(bear))
	handBefore = me.Hand.Size()
	passUntilTaxed(t, g, opp.ID)
	if g.Turn.Step != game.StepPrecombatMain {
		t.Fatalf("the question was asked in %v, want the precombat main phase", g.Turn.Step)
	}
	return g, spell, bear, handBefore
}

// openPayUnlessFor is the open pay_unless addressed to chooser, or nil.
func openPayUnlessFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// TestPayUnlessOnAResolvingSpellHoldsTheTable is #5793's report, fixed:
// with the plain PayUnless, the spell waits for the payer. The draw and
// the spell's exit come after the answer, and nothing passes and the step
// does not end until then.
func TestPayUnlessOnAResolvingSpellHoldsTheTable(t *testing.T) {
	g, spell, bear, handBefore := castRhysticReturn(t)
	me, opp := g.Seats[0], g.Seats[1]

	ask := openPayUnlessFor(g, opp.ID)
	if ask == nil || !ask.OwedInStep.IsCurrent(g.Turn) {
		t.Fatalf("the question is not owed in the step it was asked in: %+v", ask)
	}
	// The rest of the resolution waits: the draw (CR 608.2c's later
	// instruction) and the sorcery's exit (CR 608.2n).
	if got := me.Hand.Size(); got != handBefore {
		t.Errorf("hand = %d before the answer, want %d: the draw ran early", got, handBefore)
	}
	if !g.Stack.Contains(spell) || !g.ResolutionPaused() {
		t.Errorf("before the answer: spell on the stack %v, resolution paused %v; want both",
			g.Stack.Contains(spell), g.ResolutionPaused())
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Errorf("a pass with the question open = %v, want ErrChoicePending", err)
	}
	if _, err := g.AdvanceStep(); !errors.Is(err, game.ErrChoicePending) {
		t.Errorf("advance_step with the question open = %v, want ErrChoicePending", err)
	}
	if g.Turn.Step != game.StepPrecombatMain {
		t.Fatalf("the cursor left the step with the question open (now %v)", g.Turn.Step)
	}

	answerPayUnless(t, g, opp.ID, false)
	g.SettleResolution()
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("declining: the creature should be in its owner's hand")
	}
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand = %d after the answer, want %d: the draw never ran", got, handBefore+1)
	}
	if g.Stack.Contains(spell) || !me.Graveyard.Contains(spell) {
		t.Error("after the answer the spell should have left the stack for its owner's graveyard")
	}
	if g.Turn.Step != game.StepPrecombatMain {
		t.Errorf("the answer moved the cursor to %v", g.Turn.Step)
	}
}

// TestPayUnlessOnAResolvingSpellIsARestorePoint: a strict restore taken
// with the question open resumes the paused resolution, and the decline
// and the draw still run after it.
func TestPayUnlessOnAResolvingSpellIsARestorePoint(t *testing.T) {
	g, spell, bear, handBefore := castRhysticReturn(t)
	g = restoreRoundTrip(t, g, true)
	me, opp := g.Seats[0], g.Seats[1]
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("the restored table lost the question")
	}
	if !g.Stack.Contains(spell) || !g.ResolutionPaused() {
		t.Fatal("the restored table lost the paused resolution")
	}
	if _, err := g.AdvanceStep(); !errors.Is(err, game.ErrChoicePending) {
		t.Errorf("restored advance_step with the question open = %v, want ErrChoicePending", err)
	}
	answerPayUnless(t, g, opp.ID, false)
	g.SettleResolution()
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("declining after a restore: the creature should be in its owner's hand")
	}
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand = %d after the answer, want %d: the draw never ran", got, handBefore+1)
	}
	if g.Stack.Contains(spell) || !me.Graveyard.Contains(spell) {
		t.Error("after the answer the spell should have left the stack for its owner's graveyard")
	}
}

// TestPayUnlessOnAResolvingSpellPaidKeepsTheCreature: paying runs no
// decline, and the draw still follows the answer.
func TestPayUnlessOnAResolvingSpellPaidKeepsTheCreature(t *testing.T) {
	g, _, bear, handBefore := castRhysticReturn(t)
	me, opp := g.Seats[0], g.Seats[1]
	b06AddMana(opp, "C")
	answerPayUnless(t, g, opp.ID, true)
	g.SettleResolution()
	if !g.Battlefield.Contains(bear) {
		t.Error("paying {1}: the creature should stay on the battlefield")
	}
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand = %d after the payment, want %d", got, handBefore+1)
	}
}

// TestCrystalShardHoldsTheTable: the catalog's own instance of the shape.
// Crystal Shard is Rhystic Deluge's clause on an activated ability, and
// its plain PayUnless now holds the table until the target's controller
// answers.
func TestCrystalShardHoldsTheTable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shard := mrd2Catalog(g, me.ID, "Crystal Shard", "Artifact", mrdCrystalShardOracle, "{3}", 0, 0)
	bear := seedBear(g, opp.ID)
	advanceToMain(t, g)

	b06AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, shard, 1, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate the {U} row: %v", err)
	}
	passUntilTaxed(t, g, opp.ID)
	if !g.ResolutionPaused() {
		t.Error("Crystal Shard's resolution is not paused on its prompt")
	}
	step := g.Turn.Step
	if _, err := g.AdvanceStep(); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("advance_step with Crystal Shard's question open = %v, want ErrChoicePending", err)
	}
	if g.Turn.Step != step {
		t.Fatalf("the cursor moved to %v with Crystal Shard's question open", g.Turn.Step)
	}
	answerPayUnless(t, g, opp.ID, false)
	if !opp.Hand.Contains(bear) {
		t.Error("declining: the creature should be in its owner's hand")
	}
	if g.ResolutionPaused() {
		t.Error("the resolution is still paused after the answer")
	}
}
