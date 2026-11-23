package testsp

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// castRhysticReturn casts the fixture spell from seat 0 at a creature seat
// 1 controls, resolves it, and returns the table with the payer's question
// open: the spell, the creature and seat 0's hand size just before the
// spell resolved.
func castRhysticReturn(t *testing.T) (g *game.Game, spell, bear uuid.UUID, handBefore int) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear = effectstest.PushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	spell = effectstest.CastCatalogSpell(t, g, "Fixture Rhystic Return", "Sorcery", effectstest.PayUnlessSpellOracle, effectstest.CardRefs(bear))
	handBefore = me.Hand.Size()
	effectstest.PassUntilTaxed(t, g, opp.ID)
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

	effectstest.AnswerPayUnless(t, g, opp.ID, false)
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
	g = effectstest.RestoreRoundTrip(t, g, true)
	me, opp := g.Seats[0], g.Seats[1]
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("the restored table lost the question")
	}
	if !g.Stack.Contains(spell) || !g.ResolutionPaused() {
		t.Fatal("the restored table lost the paused resolution")
	}
	if _, err := g.AdvanceStep(); !errors.Is(err, game.ErrChoicePending) {
		t.Errorf("restored advance_step with the question open = %v, want ErrChoicePending", err)
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
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
	effectstest.B06AddMana(opp, "C")
	effectstest.AnswerPayUnless(t, g, opp.ID, true)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shard := effectstest.Mrd2Catalog(g, me.ID, "Crystal Shard", "Artifact", effectstest.MrdCrystalShardOracle, "{3}", 0, 0)
	bear := effectstest.SeedBear(g, opp.ID)
	effectstest.AdvanceToMain(t, g)

	effectstest.B06AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, shard, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate the {U} row: %v", err)
	}
	effectstest.PassUntilTaxed(t, g, opp.ID)
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
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	if !opp.Hand.Contains(bear) {
		t.Error("declining: the creature should be in its owner's hand")
	}
	if g.ResolutionPaused() {
		t.Error("the resolution is still paused after the answer")
	}
}
