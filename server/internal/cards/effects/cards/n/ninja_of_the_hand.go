package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ninja of the Hand — Creature — Human Ninja Villain {2}{B}, 2/2 (Marvel
// Super Heroes, #4455):
//
//	"Deathtouch
//	 Power-up — {4}{B}: Each opponent discards a card. Put a +1/+1 counter
//	 on this creature. (Activate each power-up ability only once. Reduce
//	 the cost by its mana cost if it entered this turn.)"
//
// Power-up is the engine's (CR 702.193). The effect is Tinybones's "each
// opponent discards a card" (every opponent chooses their own discard at
// once) followed by the counter, which does not wait on the discards.
//
// No simplification.
func init() {
	ab := effects.PowerUp("{4}{B}: Each opponent discards a card. Put a +1/+1 counter on this creature.",
		effects.ManaCost("{4}{B}"), func(g *game.Game, item *game.StackItem) error {
			if err := effects.EachOpponentDiscardsOne(g, item); err != nil {
				return err
			}
			return effects.PlusOneCountersOnThis(1)(g, item)
		})
	ab.Purpose.Answers = game.AnswerPump
	effects.Register(effects.Spec{
		OracleID:        "c38f7447-e646-438d-81bd-fe2a71c8a79c",
		Name:            "Ninja of the Hand",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated:       []effects.ActivatedAbility{ab},
	})
}
