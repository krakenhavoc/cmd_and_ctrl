package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Splitting Slime — Creature — Ooze {3}{G}{G}, 3/3 (Conspiracy: Take the
// Crown, S82 older sets batch 271, #5540):
//
//	"{4}{G}{G}: Monstrosity 3. (If this creature isn't monstrous, put three
//	 +1/+1 counters on it and it becomes monstrous.)
//	 When this creature becomes monstrous, create a token that's a copy of
//	 this creature. (The token has no counters and isn't monstrous.)"
//
// The shared Monstrosity activation and WhenBecomesMonstrous (CR 701.37c),
// whose body is Adagia's token copy of the Slime itself. A token copy takes
// the copiable values only (CR 707.2), so it has no counters and is not
// monstrous, and it can monstrosity on its own. The copy is of the Slime as
// it last existed if it has left by the time the trigger resolves
// (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2153f2b-4a9f-40b6-8862-0e4143686c8c",
		Name:         "Splitting Slime",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Monstrosity(effects.ManaCost("{4}{G}{G}"), 3)},
		Triggered: []game.TriggeredAbility{
			effects.WhenBecomesMonstrous("Splitting Slime — create a token that's a copy of this creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.CreateTokenCopy{Controller: ctx.Controller(), Copy: item.SourceCardID, N: 1}.Apply(ctx)
				}),
		},
	})
}
