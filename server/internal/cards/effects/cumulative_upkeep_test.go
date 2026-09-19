package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cumulative_upkeep_test.go — #567 / CR 702.24. The counter goes on,
// then the printed cost is charged once per counter, and declining
// sacrifices.

const oracleMysticRemora = "8a52f3c0-2552-4425-b2e3-5496eb2232a7"

func remoraUpkeepPrompt(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == chooser {
			return c
		}
	}
	t.Fatalf("no cumulative-upkeep prompt for %s", chooser)
	return nil
}

func ageCountersOn(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		return 0
	}
	return c.Counters[game.CounterAge]
}

// The cost scales: {1}, then {1}{1}, then {1}{1}{1} — one age counter
// added each upkeep, and the printed cost charged once per counter.
func TestCumulativeUpkeepAddsAnAgeCounterAndScalesItsCost(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	remora := pushPermanentForTest(g, owner.ID, "Mystic Remora", oracleMysticRemora, "Enchantment")

	for turn := 1; turn <= 3; turn++ {
		if turn > 1 {
			// Leave the upkeep we are standing in before looking for
			// the next one, or advanceToUpkeepOf returns immediately.
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("turn %d: leaving the previous upkeep: %v", turn, err)
			}
		}
		advanceToUpkeepOf(t, g, 1)
		if triggerOnStack(g, remora) == nil {
			t.Fatalf("turn %d: the cumulative upkeep trigger was not put on the stack", turn)
		}
		passPriorityAroundTable(t, g)
		if got := ageCountersOn(t, g, remora); got != turn {
			t.Fatalf("turn %d: age counters = %d, want %d", turn, got, turn)
		}
		choice := remoraUpkeepPrompt(t, g, owner.ID)
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
		if _, ok := battlefieldCard(g, remora); !ok {
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
			g := newCatalogGame(t)
			owner := g.Seats[1]
			remora := pushPermanentForTest(g, owner.ID, "Mystic Remora", oracleMysticRemora, "Enchantment")

			advanceToUpkeepOf(t, g, 1)
			passPriorityAroundTable(t, g)
			choice := remoraUpkeepPrompt(t, g, owner.ID)
			if err := g.ResolvePayUnless(choice.ID, owner.ID, tc.pay); err != nil {
				t.Fatalf("answer the upkeep: %v", err)
			}
			if _, ok := battlefieldCard(g, remora); ok {
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
// depends on the answer (ADR 0018 §6, PayUnless.Blocking).
func TestCumulativeUpkeepPromptBlocksTheTable(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	pushPermanentForTest(g, owner.ID, "Mystic Remora", oracleMysticRemora, "Enchantment")

	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	choice := remoraUpkeepPrompt(t, g, owner.ID)
	if !choice.ForceBlocks {
		t.Fatal("the cumulative-upkeep prompt did not ask to block the table")
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
	g := newCatalogGame(t)
	owner := g.Seats[1]
	remora := pushPermanentForTest(g, owner.ID, "Mystic Remora", oracleMysticRemora, "Enchantment")
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	choice := remoraUpkeepPrompt(t, g, owner.ID)

	clone := g.Clone()
	if err := clone.ResolvePayUnless(choice.ID, owner.ID, false); err != nil {
		t.Fatalf("answering the clone: %v", err)
	}
	if _, ok := battlefieldCard(clone, remora); ok {
		t.Error("the clone's Remora survived a declined upkeep")
	}
	if _, ok := battlefieldCard(g, remora); !ok {
		t.Error("answering the clone sacrificed the original's Remora")
	}
	// The original can still be answered the other way.
	owner.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ResolvePayUnless(choice.ID, owner.ID, true); err != nil {
		t.Fatalf("answering the original after the clone: %v", err)
	}
	if _, ok := battlefieldCard(g, remora); !ok {
		t.Error("the original's Remora was sacrificed despite being paid for")
	}
}

// The Rhystic half still fires. Seat 0 is the active player in
// newCatalogGame, so the Remora lives under seat 1 and seat 0's cast
// is an opponent's cast from its point of view.
func TestMysticRemoraTaxesOpponentNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	owner := g.Seats[1]
	remora := pushPermanentForTest(g, owner.ID, "Mystic Remora", oracleMysticRemora, "Enchantment")
	handBefore := owner.Hand.Size()

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
	if triggerOnStack(g, remora) == nil {
		t.Fatal("Mystic Remora's Rhystic trigger did not fire on an opponent's noncreature spell")
	}
	passPriorityAroundTable(t, g)

	prompt := answerPayUnless(t, g, caster.ID, false)
	if prompt.PayCost != "{4}" {
		t.Errorf("Remora's tax = %q, want {4}", prompt.PayCost)
	}
	if prompt.ForceBlocks {
		t.Error("the Rhystic half blocked the table — ADR 0018 §6 says it must not")
	}
	if got := owner.Hand.Size() - handBefore; got != 1 {
		t.Errorf("owner hand delta %d, want 1", got)
	}
}

// ...and only on a NONCREATURE spell.
func TestMysticRemoraIgnoresCreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	owner := g.Seats[1]
	remora := pushPermanentForTest(g, owner.ID, "Mystic Remora", oracleMysticRemora, "Enchantment")

	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	if triggerOnStack(g, remora) != nil {
		t.Fatal("Mystic Remora triggered on a creature spell")
	}
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, caster.ID) {
		t.Error("Mystic Remora taxed a creature spell")
	}
}
