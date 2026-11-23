package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glimpse the Future — Sorcery {2}{U} (Magic 2014, #5563):
//
//	"Look at the top three cards of your library. Put one of them into your hand and the rest into your graveyard."
//
// Wail of the Forgotten's third mode as a whole sorcery: one card to hand,
// the others put into the graveyard (not milled).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff5a2bcf-efd0-4bdb-af52-21a93447c54c",
		Name:         "Glimpse the Future",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			player := item.Controller
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 3),
				Max:    1,
				Label:  "Glimpse the Future — put one into your hand",
				Key:    effects.TakeRestGraveyard,
			}.Apply(ctx)
		},
	})
}
