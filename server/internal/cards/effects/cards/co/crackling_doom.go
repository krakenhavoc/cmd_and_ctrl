package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crackling Doom — Instant {R}{W}{B} (EDHREC rank 2122):
//
//	"Crackling Doom deals 2 damage to each opponent. Each opponent
//	 sacrifices a creature with the greatest power among creatures
//	 that player controls."
//
// Mardu's answer to the one hexproof fatty. The damage is the ping
// the pirates share; the edict is EachPlayerSacrifices with the
// GreatestPowerYouControl predicate evaluated per opponent — each
// opponent is offered exactly the creatures tied for the greatest
// power they control ("the greatest" picks out the whole tied set,
// and the player picks among it), read with counters and anthems, and an opponent with no
// creature sacrifices nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee81f37b-2a81-46d7-8d2f-9091123846c4",
		Name:         "Crackling Doom",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DamageToEachOpponent(ctx.Game, item, 2); err != nil {
				return err
			}
			return effects.EachPlayerSacrifices{
				ExceptController: true,
				Match:            effects.GreatestPowerYouControl(),
				Label:            "a creature with the greatest power among creatures you control",
			}.Apply(ctx)
		},
	})
}
