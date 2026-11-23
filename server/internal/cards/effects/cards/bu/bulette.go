package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bulette — Creature — Beast {3}{G}, 3/3 (Adventures in the Forgotten
// Realms, S82):
//
//	"At the beginning of your end step, if a creature died this turn, put a
//	 +1/+1 counter on this creature."
//
// The end-step trigger with its intervening "if" (CondAnyCreatureDied, any controller's creature, tokens included).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4e690b82-9877-41f2-9dcd-cffdb6e3cd64",
		Name:         "Bulette",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			// CR 603.4: checked as the step begins, re-checked on resolution.
			effects.On(game.EventBeginEndStep, effects.AllOf(effects.ByYou, effects.S82AfrB07AnyCreatureDiedWhen),
				"Bulette — put a +1/+1 counter on this creature",
				effects.Do(effects.If{Cond: effects.CondAnyCreatureDied{}, Then: []effects.Applier{effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}}})),
		},
	})
}
