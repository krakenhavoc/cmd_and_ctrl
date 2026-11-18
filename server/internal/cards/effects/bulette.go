package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "4e690b82-9877-41f2-9dcd-cffdb6e3cd64",
		Name:         "Bulette",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			// CR 603.4: checked as the step begins, re-checked on resolution.
			On(game.EventBeginEndStep, AllOf(ByYou, s82AfrB07AnyCreatureDiedWhen),
				"Bulette — put a +1/+1 counter on this creature",
				Do(If{Cond: CondAnyCreatureDied{}, Then: []Applier{CounterOnThis{Kind: game.CounterPlusOne, N: 1}}})),
		},
	})
}
