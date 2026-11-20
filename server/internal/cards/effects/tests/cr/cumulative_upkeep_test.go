package testscr

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The cost scales: {1}, then {1}{1}, then {1}{1}{1} — one age counter
// added each upkeep, and the printed cost charged once per counter.
func TestCumulativeUpkeepAddsAnAgeCounterAndScalesItsCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	remora := effectstest.PushPermanentForTest(g, owner.ID, "Mystic Remora", effectstest.OracleMysticRemora, "Enchantment")

	for turn := 1; turn <= 3; turn++ {
		if turn > 1 {
			// Leave the upkeep we are standing in before looking for
			// the next one, or advanceToUpkeepOf returns immediately.
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("turn %d: leaving the previous upkeep: %v", turn, err)
			}
		}
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		if effectstest.TriggerOnStack(g, remora) == nil {
			t.Fatalf("turn %d: the cumulative upkeep trigger was not put on the stack", turn)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.AgeCountersOn(t, g, remora); got != turn {
			t.Fatalf("turn %d: age counters = %d, want %d", turn, got, turn)
		}
		choice := effectstest.RemoraUpkeepPrompt(t, g, owner.ID)
		// Cost is the printed {1} once per counter: ParseCost sums the
		// repeated generic symbols.
		parsed, err := game.ParseCost(choice.PayCost)
		if err != nil {
			t.Fatalf("turn %d: cost %q does not parse: %v", turn, choice.PayCost, err)
		}
		if parsed.Generic != turn {
			t.Fatalf("turn %d: cost %q is {%d} generic, want {%d}", turn, choice.PayCost, parsed.Generic, turn)
		}
		for i := 0; i < turn; i++ {
			owner.ManaPool.AddMana(game.ManaToken{Color: "C"})
		}
		if err := g.ResolvePayUnless(choice.ID, owner.ID, true); err != nil {
			t.Fatalf("turn %d: pay the upkeep: %v", turn, err)
		}
		if _, ok := effectstest.BattlefieldCard(g, remora); !ok {
			t.Fatalf("turn %d: a paid cumulative upkeep sacrificed the permanent", turn)
		}
	}
}

// Declining sacrifices it (CR 702.24a), and so does a "pay" the
// controller cannot fund — the pay-unless contract.
func TestCumulativeUpkeepSacrificesOnDecline(t *testing.T) {
	for _, tc := range []struct {
		name string
		pay  bool
		fund bool
	}{
		{"declined", false, false},
		{"paid without the mana", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			owner := g.Seats[1]
			remora := effectstest.PushPermanentForTest(g, owner.ID, "Mystic Remora", effectstest.OracleMysticRemora, "Enchantment")

			effectstest.AdvanceToUpkeepOf(t, g, 1)
			effectstest.PassPriorityAroundTable(t, g)
			choice := effectstest.RemoraUpkeepPrompt(t, g, owner.ID)
			if err := g.ResolvePayUnless(choice.ID, owner.ID, tc.pay); err != nil {
				t.Fatalf("answer the upkeep: %v", err)
			}
			if _, ok := effectstest.BattlefieldCard(g, remora); ok {
				t.Error("Mystic Remora survived an unpaid cumulative upkeep")
			}
			if owner.Graveyard == nil || owner.Graveyard.Size() == 0 {
				t.Error("the sacrificed permanent did not reach its owner's graveyard")
			}
		})
	}
}

// The prompt BLOCKS the table, unlike every other pay_unless: the
// question is the active player's own and the rest of their turn
// depends on the answer (ADR 0018 §6, effects.UpkeepPayUnless).
func TestCumulativeUpkeepPromptBlocksTheTable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	effectstest.PushPermanentForTest(g, owner.ID, "Mystic Remora", effectstest.OracleMysticRemora, "Enchantment")

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	choice := effectstest.RemoraUpkeepPrompt(t, g, owner.ID)
	// Anchored to the upkeep it was asked in, and blocking because the
	// cursor is still standing there (#997). Nothing on the card says
	// "block": the engine reads the step.
	if choice.OwedInStep.Step != game.StepUpkeep || choice.OwedInStep.Turn != g.Turn.Seq {
		t.Fatalf("the cumulative-upkeep prompt is anchored to %v, want turn %d's upkeep",
			choice.OwedInStep, g.Turn.Seq)
	}
	if !g.ChoicePromptBlocksTable(choice) {
		t.Fatal("ChoicePromptBlocksTable says the cumulative-upkeep prompt does not block")
	}
	if _, err := g.AdvanceStep(); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("advance_step past the cumulative-upkeep prompt: %v, want ErrChoicePending", err)
	}
	// And the kind itself is untouched: Rhystic Study still does not
	// block, which is what ADR 0018 §6 decided.
	if game.ChoiceBlocksTable(game.PendingChoicePayUnless) {
		t.Error("the pay_unless KIND now blocks — this change is per prompt, not per kind")
	}
}

// Undo restores a clone, and the prompt's continuation has to resolve
// against that one.
func TestCumulativeUpkeepPromptSurvivesAClone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	remora := effectstest.PushPermanentForTest(g, owner.ID, "Mystic Remora", effectstest.OracleMysticRemora, "Enchantment")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	choice := effectstest.RemoraUpkeepPrompt(t, g, owner.ID)

	clone := g.Clone()
	if err := clone.ResolvePayUnless(choice.ID, owner.ID, false); err != nil {
		t.Fatalf("answering the clone: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(clone, remora); ok {
		t.Error("the clone's Remora survived a declined upkeep")
	}
	if _, ok := effectstest.BattlefieldCard(g, remora); !ok {
		t.Error("answering the clone sacrificed the original's Remora")
	}
	// The original can still be answered the other way.
	owner.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ResolvePayUnless(choice.ID, owner.ID, true); err != nil {
		t.Fatalf("answering the original after the clone: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, remora); !ok {
		t.Error("the original's Remora was sacrificed despite being paid for")
	}
}

// The Rhystic half still fires. Seat 0 is the active player in
// newCatalogGame, so the Remora lives under seat 1 and seat 0's cast
// is an opponent's cast from its point of view.
func TestMysticRemoraTaxesOpponentNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[0]
	owner := g.Seats[1]
	remora := effectstest.PushPermanentForTest(g, owner.ID, "Mystic Remora", effectstest.OracleMysticRemora, "Enchantment")
	handBefore := owner.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
	if effectstest.TriggerOnStack(g, remora) == nil {
		t.Fatal("Mystic Remora's Rhystic trigger did not fire on an opponent's noncreature spell")
	}
	effectstest.PassPriorityAroundTable(t, g)

	prompt := effectstest.AnswerPayUnless(t, g, caster.ID, false)
	if prompt.PayCost != "{4}" {
		t.Errorf("Remora's tax = %q, want {4}", prompt.PayCost)
	}
	if prompt.OwedInStep.NamesAStep() {
		t.Errorf("the Rhystic half anchored itself to %v and blocks the table — "+
			"ADR 0018 §6 says it must not", prompt.OwedInStep)
	}
	// "You may draw": a question, asked after the decline.
	if got := owner.Hand.Size() - handBefore; got != 0 {
		t.Errorf("drew %d before being asked", got)
	}
	effectstest.AnswerMayChoice(t, g, owner.ID, true)
	if got := owner.Hand.Size() - handBefore; got != 1 {
		t.Errorf("owner hand delta %d, want 1", got)
	}
}

// The draw is optional (the old caveat): declining it, with a library
// that would kill, keeps the controller alive and the hand unchanged.
func TestMysticRemoraDrawIsOptional(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[0]
	owner := g.Seats[1]
	effectstest.PushPermanentForTest(g, owner.ID, "Mystic Remora", effectstest.OracleMysticRemora, "Enchantment")
	owner.Library.Cards = nil
	handBefore := owner.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, caster.ID, false)
	effectstest.AnswerMayChoice(t, g, owner.ID, false)
	if owner.Hand.Size() != handBefore || owner.AttemptedEmptyDraw {
		t.Errorf("declined draw: hand %d -> %d, attempted empty draw %v", handBefore, owner.Hand.Size(), owner.AttemptedEmptyDraw)
	}
	// And a paid tax asks nothing.
	if effectstest.LatestConfirmFor(g, owner.ID) != nil {
		t.Error("a second question is open")
	}
}

// ...and only on a NONCREATURE spell.
func TestMysticRemoraIgnoresCreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[0]
	owner := g.Seats[1]
	remora := effectstest.PushPermanentForTest(g, owner.ID, "Mystic Remora", effectstest.OracleMysticRemora, "Enchantment")

	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	if effectstest.TriggerOnStack(g, remora) != nil {
		t.Fatal("Mystic Remora triggered on a creature spell")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, caster.ID) {
		t.Error("Mystic Remora taxed a creature spell")
	}
}
