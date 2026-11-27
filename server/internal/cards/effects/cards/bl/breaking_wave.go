package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Breaking Wave — Sorcery {2}{U}{U} (Invasion, S82 older sets batch 662, #5616):
//
//	"You may cast this spell as though it had flash if you pay {2} more to
//	 cast it. (You may cast it any time you could cast an instant.)
//	 Simultaneously untap all tapped creatures and tap all untapped
//	 creatures."
//
// Rout's FlashForMore {2} (CR 601.3c). The swap reads both sets of creatures
// before anything moves, then untaps the tapped and taps the untapped, so no
// creature is moved twice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "f455b731-6ff9-4acb-9447-f524faa49b27",
		Name:          "Breaking Wave",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.FlashForMore("{2}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.OldB662SwapTappedCreatures(ctx.Game, item)
		},
	})
}
