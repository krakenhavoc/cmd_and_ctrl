package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// upkeep_pay_unless_halt_test.go — #997 at the cards.
//
// The bug the issue reports is not visible in the Stasis tests beside
// this file, and that is the point of it: those answer the prompt
// immediately, which happens to dodge the interleaving. The bug needs
// the TABLE to move while the prompt is outstanding — a pass from any
// seat, or an advance_step — which is the ordinary case at a table
// with more than two seats. All four seats of newCatalogGame are used
// on purpose.
//
// Both cards ask the same sentence and both were wrong the same way
// before this: "at the beginning of your upkeep, pay <cost> or <lose
// something you cannot get back>", asked of the active player in their
// own upkeep, with the table free to walk on past it.

// TestStasisUpkeepPromptHaltsTheTurn is #997's headline, card-side.
func TestStasisUpkeepPromptHaltsTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0].ID
	stasisID := seedReplacementPermanent(g, stasisOracle, "Stasis", p0)

	walkToUpkeepOfSeat0(t, g)
	passPriorityAroundTable(t, g)
	prompt := upkeepPromptFor(t, g, p0)

	// Anchored to the upkeep it was asked in. Nothing on the card says
	// "block": the engine reads the cursor (#997).
	if prompt.OwedInStep.Step != game.StepUpkeep || prompt.OwedInStep.Turn != g.Turn.Number {
		t.Fatalf("Stasis's prompt is anchored to %v, want turn %d's upkeep",
			prompt.OwedInStep, g.Turn.Number)
	}

	// The moves the bug was: somebody passes, or the step is advanced.
	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("a pass while Stasis's upkeep prompt was open = %v, want ErrChoicePending", err)
	}
	if _, err := g.AdvanceStep(); !errors.Is(err, game.ErrChoicePending) {
		t.Errorf("advance_step while Stasis's upkeep prompt was open = %v, want ErrChoicePending", err)
	}
	if err := g.PassTurn(); !errors.Is(err, game.ErrChoicePending) {
		t.Errorf("pass_turn while Stasis's upkeep prompt was open = %v, want ErrChoicePending", err)
	}
	if g.Turn.Step != game.StepUpkeep {
		t.Fatalf("the table reached %q with the {U} unpaid", g.Turn.Step)
	}

	// And it is still settleable, which is what makes the halt a halt
	// rather than a wedge: declining sacrifices Stasis and the table
	// plays on.
	answerPayUnless(t, g, p0, false)
	if g.Battlefield.Contains(stasisID) {
		t.Error("declining the upkeep did not sacrifice Stasis")
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Errorf("advance_step after the upkeep prompt was answered: %v", err)
	}
}

// Pact of Negation is the other half of the same judgement, and the
// consequence is the largest one in the game: the table could take the
// turn — the Pact's controller could untap, draw and attack — with the
// {3}{U}{U} still owed and the loss still pending.
func TestPactOfNegationUpkeepDebtHaltsTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	victim := castCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	pact := handCardFull(me, "Pact of Negation", "Instant", "{0}", pactOfNegationOracle, []string{"U"})
	if err := g.CastSpell(me.ID, pact, game.CastSpellParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("casting Pact of Negation: %v", err)
	}
	passPriorityAroundTable(t, g)

	walkToUpkeepOfSeat0(t, g)
	passPriorityAroundTable(t, g)
	prompt := upkeepPromptFor(t, g, me.ID)
	if prompt.OwedInStep.Step != game.StepUpkeep || prompt.OwedInStep.Turn != g.Turn.Number {
		t.Fatalf("the Pact's debt is anchored to %v, want turn %d's upkeep",
			prompt.OwedInStep, g.Turn.Number)
	}

	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("a pass while the Pact's debt was open = %v, want ErrChoicePending", err)
	}
	if _, err := g.AdvanceStep(); !errors.Is(err, game.ErrChoicePending) {
		t.Errorf("advance_step while the Pact's debt was open = %v, want ErrChoicePending", err)
	}

	// Declining is the whole point of the card's downside.
	answerPayUnless(t, g, me.ID, false)
	if !me.Eliminated {
		t.Error("an unpaid Pact of Negation did not lose its controller the game")
	}
}

// upkeepPromptFor returns the open pay-unless `chooser` owes, without
// answering it — answerPayUnless answers, and these tests need the
// prompt to still be sitting there while the table tries to move.
func upkeepPromptFor(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == chooser {
			return c
		}
	}
	t.Fatalf("no upkeep pay-unless prompt for %s", chooser)
	return nil
}
