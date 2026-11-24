package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bitter Revelation — Sorcery {3}{B} (Commander Legends):
//
//	"Look at the top four cards of your library. Put two of them into your
//	 hand and the rest into your graveyard. You lose 2 life."
//
// Dark Bargain's look with four cards: a private look, a mandatory pick of
// two for the hand, the rest put into the graveyard (not milled), then the
// life loss, in the take's registered continuation so it waits for the pick.
// With fewer cards in the library you take what is there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f3e025ce-6114-42d0-842f-73366c853657",
		Name:         "Bitter Revelation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 4),
				Max:    2,
				Label:  "Bitter Revelation — put two of them into your hand and the rest into your graveyard",
				Key:    bitterRevelationTake,
			}.Apply(ctx)
		},
	})
}

// bitterRevelationTake is the look's continuation: the rest into the
// graveyard, then "You lose 2 life". An on-disk key.
var bitterRevelationTake = effects.RegisterTake("choose-cards/bitter-revelation-take",
	effects.TakeDef{Then: func(g *game.Game, res effects.TakeFromLibraryResult) error {
		if err := effects.TakeRestIntoGraveyard(g, res); err != nil {
			return err
		}
		return g.ChangePlayerLifeForEffect(res.Source, res.Player, -2)
	}})
