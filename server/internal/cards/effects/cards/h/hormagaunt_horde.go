package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hormagaunt Horde — Creature — Tyranid {X}{G}, 1/1:
//
//	"Ravenous (This creature enters with X +1/+1 counters on it. If X is 5 or
//	 more, draw a card when it enters.)
//	 Endless Swarm — Whenever a land you control enters, you may pay {2}{G}.
//	 If you do, return this card from your graveyard to your hand."
//
// Silvan Reveler's shape: a landfall trigger that works from the graveyard and
// offers the payment as it resolves, for the same graveyard object (CR 400.7).
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "cc965557-f69e-4689-8791-d3618e4ffbfa",
		Name:                       "Hormagaunt Horde",
		Completeness:               effects.CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: effects.RavenousEntry(),
		Triggered: []game.TriggeredAbility{
			effects.RavenousDraw("Hormagaunt Horde"),
			effects.InGraveyard(effects.Landfall("Hormagaunt Horde — you may pay {2}{G} to return it to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPayForGraveyardReturnToHand(g, item, "{2}{G}", "Hormagaunt Horde — pay {2}{G} to return it to your hand?")
				})),
		},
	})
}
