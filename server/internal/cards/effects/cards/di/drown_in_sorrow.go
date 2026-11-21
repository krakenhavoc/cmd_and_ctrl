package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drown in Sorrow — Sorcery {1}{B}{B} (Born of the Gods,
// #5559):
//
//	"All creatures get -2/-2 until end of turn. Scry 1. (Look at the top
//	 card of your library. You may put that card on the bottom.)"
//
// Cry of the Carnarium's shrink (BoostUntilEOT over every creature),
// then the scry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5c93ec85-f87f-4121-9fbf-61fd598d1e65",
		Name:         "Drown in Sorrow",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2},
			Scry:  1,
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.BoostUntilEOT{Match: effects.Creature(), Power: -2, Toughness: -2, Label: "Drown in Sorrow — -2/-2"}).Apply(ctx); err != nil {
				return err
			}
			return effects.Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
