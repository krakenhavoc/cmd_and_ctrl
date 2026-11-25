package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cleansing Wildfire — Sorcery {1}{R} (Zendikar Rising, S82 OLD #5508):
//
//	"Destroy target land. Its controller may search their library for a
//	 basic land card, put it onto the battlefield tapped, then shuffle.
//	 Draw a card."
//
// The land's controller (not necessarily the caster) searches, as Avengers
// Disassembled's land mode does. The caster draws whether or not the land
// was destroyed; the spell does nothing at all if its one target is
// illegal as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "993b92f4-4160-4171-9e67-90dc935030fe",
		Name:         "Cleansing Wildfire",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := effects.DestroyThenItsControllerMaySearchForATappedBasic(ctx, id,
				"Cleansing Wildfire — you may search for a basic land"); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
