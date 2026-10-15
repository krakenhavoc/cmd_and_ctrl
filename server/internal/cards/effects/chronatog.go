package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chronatog — Creature — Atog {1}{U}, 1/2:
//
//	"{0}: This creature gets +3/+3 until end of turn. You skip your next
//	 turn. Activate only once each turn."
//
// A free pump that costs a turn (CR 614.10, ADR 0188 Decision 1, #3565).
// The two sentences are separate instructions, so the skip is written
// even when Chronatog left the battlefield in response and the pump did
// nothing. "Once each turn" counts activations (CR 602.1b) through
// OncePerTurnActivation, so a second attempt is never offered.
//
// No simplification.
func init() {
	const label = "{0}: This creature gets +3/+3 until end of turn. You skip your next turn. Activate only once each turn."
	Register(Spec{
		OracleID:     "bb300e38-32d7-4152-9a82-88d43af99f19",
		Name:         "Chronatog",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump, SkipsTurns: 1},
			Cost:      ManaCost("{0}"),
			Condition: OncePerTurnActivation(label),
			Effect:    chronatogPumpAndSkip("Chronatog — +3/+3"),
		}},
	})
}

// chronatogPumpAndSkip is the shared effect of Chronatog and Chronatog
// Totem: the source gets +3/+3 until end of turn, then its controller
// skips their next turn. It captures only the label, a plain string
// (undo resolves an ability against a cloned game).
func chronatogPumpAndSkip(label string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if err := thisGetsUntilEndOfTurn(3, 3, label)(g, item); err != nil {
			return err
		}
		return SkipYourNextTurns{}.Apply(NewContext(g, item))
	}
}
