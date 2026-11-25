package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geomancer's Gambit — Sorcery {2}{R} (Modern Horizons):
//
//	"Destroy target land. Its controller may search their library for a basic
//	 land card, put it onto the battlefield, then shuffle.
//	 Draw a card."
//
// Cleansing Wildfire with the basic land entering untapped: the land's
// controller (not necessarily the caster) searches, and the caster draws
// whether or not the land was destroyed. The spell does nothing at all if its
// one target is illegal as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33bb4aff-6476-4787-b3fe-028b1b1bd1bc",
		Name:         "Geomancer's Gambit",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := effects.OldB159DestroyThenItsControllerMaySearchForABasic(ctx, id,
				"Geomancer's Gambit — you may search for a basic land"); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
