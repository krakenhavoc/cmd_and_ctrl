package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thoughtweft Gambit — Instant {4}{W/U}{W/U} (Shadowmoor, S82 older sets):
//
//	"Tap all creatures your opponents control and untap all creatures you
//	 control."
//
// Gideon, Martial Paragon's two mass effects in the printed order: the
// opponents' creatures are tapped, then yours are untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ddf8d3d2-b0f9-43bc-9c9e-cb08fdedcc72",
		Name:         "Thoughtweft Gambit",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.TapAllCreaturesYourOpponentsControl(ctx.Game, item); err != nil {
				return err
			}
			return effects.UntapAllCreaturesYouControl{}.Apply(ctx)
		},
	})
}
