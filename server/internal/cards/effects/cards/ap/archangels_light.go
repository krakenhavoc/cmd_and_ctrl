package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Archangel's Light — Sorcery {7}{W}:
//
//	"You gain 2 life for each card in your graveyard, then shuffle your
//	 graveyard into your library."
//
// The count is read as the spell resolves, with the spell itself still on the
// stack, so it is not one of the cards counted.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc734a75-93bf-448b-b347-810ae835fe49",
		Name:         "Archangel's Light",
		Purpose:      game.Purpose{LifeGain: 2},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			me := ctx.Controller()
			if p := ctx.PlayerByID(me); p != nil && p.Graveyard != nil {
				if err := (effects.GainLife{Amount: 2 * p.Graveyard.Size()}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.FinaleShuffleGraveyardIntoLibrary(ctx, me)
		},
	})
}
