package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ransack the Lab — Sorcery {1}{B} (Modern Horizons, #5519):
//
//	"Look at the top three cards of your library. Put one of them into your
//	 hand and the rest into your graveyard."
//
// Resentful Revelation's body without the flashback: take one of the top
// three, the rest go to the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9b97a7fe-c5d9-4698-a49c-183b3f1ce4db",
		Name:         "Ransack the Lab",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 3),
				Max:    1,
				Label:  "Ransack the Lab — put one of them into your hand and the rest into your graveyard",
				Key:    effects.TakeRestGraveyard,
			}.Apply(ctx)
		},
	})
}
