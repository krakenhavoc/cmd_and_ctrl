package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// skip_step_pairs_test.go is the catalog half of #710: TWO skip-step
// replacements applying to one step transition.
//
// One of them was always fine — the apply-loop's single-applicable
// fast path cancels the event and the cursor walks past the step. Two
// of them made CR 616 ask the affected player to order the pair, and
// the step-transition resume was a stub, so the prompt was queued and
// the step then ran anyway. A player with Necropotence AND Yawgmoth's
// Bargain drew a card every turn; a table with two Stasis untapped.
//
// Both effects are pure cancels (ReplacementEffect.PureCancel), so
// there is no ordering to ask about and no prompt is queued at all.
// The engine-side coverage of the pause itself — the mixed case where
// the prompt is real — lives in
// server/internal/game/step_transition_resume_test.go.

// TestNecropotenceAndYawgmothsBargainSkipTheDrawStepOnce — the issue's
// headline pair. Both print "skip your draw step", both cancel the
// same transition, and one skipped step is the whole of what they do
// together.
func TestNecropotenceAndYawgmothsBargainSkipTheDrawStepOnce(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	_ = seedReplacementPermanent(g, necropotenceOracle, "Necropotence", owner.ID)
	_ = seedReplacementPermanent(g, yawgmothsBargainOracle, "Yawgmoth's Bargain", owner.ID)

	handBefore := owner.Hand.Size()
	trace := walkSteps(t, g, 200, func() bool {
		return g.Turn.ActiveSeat == 2 && g.Turn.Step == game.StepPrecombatMain
	})

	if n := orderPrompts(g); n != 0 {
		t.Errorf("%d CR 616 ordering prompts queued, want 0 — two pure cancels have "+
			"nothing to order", n)
	}
	if sawStep(trace, 1, game.StepDraw) {
		t.Error("seat 1's draw step happened despite Necropotence AND Yawgmoth's Bargain (#710)")
	}
	if owner.Hand.Size() != handBefore {
		t.Errorf("seat 1 drew %d cards with two skip-your-draw-step effects in play (#710)",
			owner.Hand.Size()-handBefore)
	}
	if !sawStep(trace, 2, game.StepDraw) {
		t.Error("seat 2's draw step was skipped too — the replacement is not seat-scoped")
	}
}

// TestTwoStasisSkipTheUntapStepOnce — the same shape on the other
// catalog step skip, and the second half of #710's repro. Stasis is
// not legendary, so two of them on the battlefield is a legal board;
// two untap skips are still one skipped untap step, and a permanent
// that was tapped stays tapped.
func TestTwoStasisSkipTheUntapStepOnce(t *testing.T) {
	g := newCatalogGame(t)
	p0 := g.Seats[0].ID
	_ = seedReplacementPermanent(g, stasisOracle, "Stasis", p0)
	_ = seedReplacementPermanent(g, stasisOracle, "Stasis", p0)

	bearID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: bearID,
		Name:       "Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      p0,
		Controller: p0,
		Tapped:     true,
	})

	trace := walkSteps(t, g, 200, func() bool {
		return g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepUpkeep && g.Turn.Number >= 2
	})

	// Only ordering prompts are counted: two Stasis also queue their
	// two "sacrifice unless you pay {U}" upkeep triggers, which is
	// the card working, not #710.
	if n := orderPrompts(g); n != 0 {
		t.Errorf("%d CR 616 ordering prompts queued, want 0 — two untap skips are "+
			"interchangeable", n)
	}
	if sawStep(trace, 0, game.StepUntap) {
		t.Error("seat 0's untap step happened with two Stasis on the battlefield (#710)")
	}
	var tapped bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == bearID {
				tapped = c.Tapped
				return
			}
		}
	})
	if !tapped {
		t.Error("Bears untapped with two Stasis skipping the untap step (#710)")
	}
}

// orderPrompts counts the CR 616 replacement-ordering prompts sitting
// in the queue.
func orderPrompts(g *game.Game) int {
	var n int
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceReplacementOrder {
				n++
			}
		}
	})
	return n
}
