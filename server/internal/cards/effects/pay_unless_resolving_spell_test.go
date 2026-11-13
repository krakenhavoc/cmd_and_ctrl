package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pay_unless_resolving_spell_test.go — #5793: does a PayUnless asked by a
// RESOLVING SPELL hold the table while "that player" decides?
//
// The spell below is Rhystic Deluge's and Withdraw's clause on a sorcery:
// "Return target creature to its owner's hand unless its controller pays
// {1}. Draw a card." The draw is a later instruction (CR 608.2c), so it
// shows whether the rest of the resolution waits for the answer.
//
// Two copies of it differ in one field of the prompt:
//
//   - the plain one asks through effects.PayUnless, the primitive the
//     Prophecy agent reached for;
//   - the held one asks through the same keyed door with InThisStep set,
//     the anchor AnyPlayerPays and UpkeepPayUnless use (OwedInStep, #997,
//     ADR 0150 decision 2).
//
// What the tests record is the engine as it stands. A pay_unless KIND does
// not stop the table (ADR 0018 §6, choice_gate.go), and nothing derives a
// halt for this shape, so the plain prompt is asked after the spell has
// already finished: the draw happens first, the sorcery is in the graveyard,
// any seat may pass and the step may end, and the bounce lands whenever the
// payer answers. The held prompt is a resolution pause (resolution_pause.go):
// the draw and the spell's exit wait for the answer (#4012, #3613), and no
// pass or advance gets through until then. Both survive a strict restore
// with the question open.
//
// If an owner decision makes the plain shape hold (#5793), the first test is
// the one to flip.

const (
	payUnlessSpellOracle     = "test-5793-pay-unless-spell"
	payUnlessHeldSpellOracle = "test-5793-pay-unless-held-spell"
)

func init() {
	Register(Spec{
		OracleID: payUnlessSpellOracle,
		Name:     "Fixture Rhystic Return",
		Targets:  TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return askTheTargetsControllerForOne(ctx, false) }),
			DrawCards{N: 1},
		),
	})
	Register(Spec{
		OracleID: payUnlessHeldSpellOracle,
		Name:     "Fixture Held Rhystic Return",
		Targets:  TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return askTheTargetsControllerForOne(ctx, true) }),
			DrawCards{N: 1},
		),
	})
}

// askTheTargetsControllerForOne asks the target creature's controller for
// {1}; the decline is Crystal Shard's registered bounce, so no test-only
// key is added to the ledger. held anchors the prompt to the current step.
func askTheTargetsControllerForOne(ctx *Context, held bool) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		question := "pay {1}, or " + c.Name + " returns to its owner's hand?"
		if !held {
			return PayUnless{Chooser: c.Controller, Cost: "{1}", Question: question, Else: crystalShardBounce}.Apply(ctx)
		}
		return ctx.Game.QueuePayUnlessPromptKeyedForEffect(game.PayUnlessKeyedPrompt{
			Chooser: c.Controller, Source: ctx.Source(), Cost: "{1}", Question: question,
			Then: crystalShardBounce, Carry: []uuid.UUID{ctx.Controller()},
			Asked: game.NewAskedItem(ctx.Item), InThisStep: true,
		})
	}
	return nil
}

// castRhysticReturn casts one of the fixture spells from seat 0 at a
// creature seat 1 controls, resolves it, and returns the table with the
// payer's question open: the spell, the creature and seat 0's hand size
// just before the spell resolved.
func castRhysticReturn(t *testing.T, oracle string) (g *game.Game, spell, bear uuid.UUID, handBefore int) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear = pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	spell = castCatalogSpell(t, g, "Fixture Rhystic Return", "Sorcery", oracle, cardRefs(bear))
	handBefore = me.Hand.Size()
	passUntilTaxed(t, g, opp.ID)
	if g.Turn.Step != game.StepPrecombatMain {
		t.Fatalf("the question was asked in %v, want the precombat main phase", g.Turn.Step)
	}
	return g, spell, bear, handBefore
}

// TestPayUnlessOnAResolvingSpellDoesNotHoldTheTable is #5793's report,
// reproduced: with the plain PayUnless, the spell finishes before the payer
// answers, and the table plays on with the question open.
func TestPayUnlessOnAResolvingSpellDoesNotHoldTheTable(t *testing.T) {
	g, spell, bear, handBefore := castRhysticReturn(t, payUnlessSpellOracle)
	me, opp := g.Seats[0], g.Seats[1]

	// The rest of the resolution ran early: the draw (CR 608.2c's later
	// instruction) and the sorcery's exit (CR 608.2n) are already done.
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand = %d before the answer, want %d: the draw waited for it", got, handBefore+1)
	}
	if g.Stack.Contains(spell) || !me.Graveyard.Contains(spell) {
		t.Errorf("before the answer the spell is on the stack %v, in the graveyard %v; want gone from the stack",
			g.Stack.Contains(spell), me.Graveyard.Contains(spell))
	}
	if g.ResolutionPaused() {
		t.Error("the resolution is paused on the plain prompt")
	}

	// Priority passes and the step ends with the question open.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("a pass with the question open = %v, want it to go through", err)
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("advance_step with the question open = %v, want it to go through", err)
	}
	if g.Turn.Step == game.StepPrecombatMain {
		t.Fatal("the cursor is still in the step the question was asked in")
	}
	if !hasPayUnlessFor(g, opp.ID) || !g.Battlefield.Contains(bear) {
		t.Fatal("the question should still be open and the creature still on the battlefield")
	}

	// The effect lands late, in a later step, when the payer answers.
	answerPayUnless(t, g, opp.ID, false)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("declining late: the creature should be in its owner's hand")
	}
}

// TestPayUnlessOnAResolvingSpellIsARestorePoint: the plain prompt survives
// a strict restore with the question open, and the decline still runs.
func TestPayUnlessOnAResolvingSpellIsARestorePoint(t *testing.T) {
	g, _, bear, _ := castRhysticReturn(t, payUnlessSpellOracle)
	g = restoreRoundTrip(t, g, true)
	opp := g.Seats[1]
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("the restored table lost the question")
	}
	answerPayUnless(t, g, opp.ID, false)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("declining after a restore: the creature should be in its owner's hand")
	}
}

// TestStepHeldPayUnlessOnAResolvingSpellHoldsTheTable is the comparison:
// the same prompt anchored to its step (the seam AnyPlayerPays and
// UpkeepPayUnless use) pauses the resolution, so nothing moves and the rest
// of the spell waits for the answer.
func TestStepHeldPayUnlessOnAResolvingSpellHoldsTheTable(t *testing.T) {
	g, spell, bear, handBefore := castRhysticReturn(t, payUnlessHeldSpellOracle)
	me, opp := g.Seats[0], g.Seats[1]

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

	// A restore mid-question resumes the paused resolution.
	g = restoreRoundTrip(t, g, true)
	me, opp = g.Seats[0], g.Seats[1]
	if !g.Stack.Contains(spell) || !g.ResolutionPaused() {
		t.Fatal("the restored table lost the paused resolution")
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

// TestCrystalShardDoesNotHoldTheTable: the catalog's own instance of the
// shape. Crystal Shard is Rhystic Deluge's clause on an activated ability,
// and its prompt is the plain PayUnless, so the table plays on with it open.
func TestCrystalShardDoesNotHoldTheTable(t *testing.T) {
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
	if g.ResolutionPaused() {
		t.Error("Crystal Shard's resolution is paused on its prompt")
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("advance_step with Crystal Shard's question open = %v, want it to go through", err)
	}
	answerPayUnless(t, g, opp.ID, false)
	if !opp.Hand.Contains(bear) {
		t.Error("declining late: the creature should be in its owner's hand")
	}
}
