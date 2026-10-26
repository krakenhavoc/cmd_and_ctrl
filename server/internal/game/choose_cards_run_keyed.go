package game

import (
	"fmt"

	"github.com/google/uuid"
)

// choose_cards_run_keyed.go — a choose_cards run whose continuation and
// legs are KEYED (#4290, ADR 0013 §5x / §5y amendment of 2026-10-26), so a
// table waiting on any seat's pick is a restore point.
//
// choose_cards_run.go's ChooseCardsRunThenForEffect gives every leg a
// closure: its own Then, which settles the run, and a Validate for a set
// rule. A keyed leg is a keyed choose_cards pick (choose_cards_keyed.go)
// whose registered continuation settles the run, and whose set rule, when
// it has one, is registered under the same key
// (RegisterChooseCardsRunLegRule). The run's own continuation is the usual
// registered one (RegisterPromptRunThen), handed the answer.

// chooseCardsRunLegFn is the engine's continuation for a leg of a keyed
// choose_cards run: the pick is the seat's whole answer, so the leg only
// settles the run with it. An unanswered leg settles with nothing, as the
// departure table does.
func chooseCardsRunLegFn(g *Game, r CardsChosen) error {
	if !r.Answered {
		return g.settleRunLegLocked(r.Run, r.Chooser, nil)
	}
	return g.settleRunLegLocked(r.Run, r.Chooser, r.Picked)
}

// ChooseCardsRunLeg is the continuation of a keyed run's leg that has no
// set rule of its own.
var ChooseCardsRunLeg = RegisterChooseCardsThen("choose-cards/run-leg", chooseCardsRunLegFn)

// RegisterChooseCardsRunLegRule registers the continuation of a keyed
// run's leg that has a set rule over what the seat picks (Keep creatures'
// "total power 4 or less", Explore the Vastlands' "a land and/or an
// instant or sorcery"), under its own key. The key is an on-disk
// identity, ledgered as "choose <key>". The rule is never called for an
// empty pick, is a function of the cards alone, and must capture nothing.
func RegisterChooseCardsRunLegRule(key string, rule SetRuleFunc) ChooseCardsThen {
	return RegisterChooseCardsThenRule(key, rule, chooseCardsRunLegFn)
}

// ChooseCardsRunThenKeyedForEffect is ChooseCardsRunThenForEffect with the
// rest of the card as a registered run continuation (RegisterPromptRunThen)
// and `data` the plain data it is handed, and every leg a KEYED pick, so a
// table waiting on any leg is a restore point (#4290).
//
// A leg's ThenKey is ChooseCardsRunLeg unless it names a leg with a set
// rule (RegisterChooseCardsRunLegRule). The legs hold no closure: Then and
// Validate must be nil, and Zone, which a keyed pick has no frame for, is
// read as the live-zone re-check (ChooseCardsPrompt.LiveZone) the keyed
// pick does have. The continuation reads RunDone.Picks().
//
// Caller must hold g.mu.
func (g *Game) ChooseCardsRunThenKeyedForEffect(legs []ChooseCardsPrompt, then PromptRunThen, source uuid.UUID, data RunData) (int, error) {
	asks := make([]uuid.UUID, len(legs))
	for i, leg := range legs {
		if leg.Then != nil || leg.Validate != nil {
			return 0, fmt.Errorf("%w: keyed choose_cards run %q has a leg with a closure; register its set rule (RegisterChooseCardsRunLegRule)", ErrInvalidParam, then.key)
		}
		asks[i] = leg.Chooser
	}
	// runPromptsKeyedLocked calls queue once per ask, in ask order, so the
	// next leg is the one at `next`.
	next := 0
	return g.runPromptsKeyedLocked(asks, then, source, data, func(seat, run uuid.UUID) bool {
		leg := legs[next]
		next++
		if len(leg.Cards) == 0 {
			return false
		}
		if leg.ThenKey.key == "" {
			leg.ThenKey = ChooseCardsRunLeg
		}
		if leg.LiveZone == "" {
			leg.LiveZone = leg.Zone
		}
		leg.promptRun = run
		return g.QueueChooseCardsForEffect(leg) != uuid.Nil
	})
}
