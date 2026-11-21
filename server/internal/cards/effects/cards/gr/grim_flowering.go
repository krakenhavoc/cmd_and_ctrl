package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grim Flowering — Sorcery {5}{G} (Dark Ascension, #5571):
//
//	"Draw a card for each creature card in your graveyard."
//
// The creature cards are counted once as the spell resolves (CR 608.2h),
// Nature's Resurgence's count for the caster alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "65428247-b386-410c-91b7-81dd278c0927",
		Name:         "Grim Flowering",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{N: effects.B11CreatureCardsInGraveyard(ctx.Game, ctx.Controller())}.Apply(ctx)
		},
	})
}
