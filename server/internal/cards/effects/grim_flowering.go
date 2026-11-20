package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grim Flowering — Sorcery {5}{G} (Dark Ascension, #5571):
//
//	"Draw a card for each creature card in your graveyard."
//
// The creature cards are counted once as the spell resolves (CR 608.2h),
// Nature's Resurgence's count for the caster alone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "65428247-b386-410c-91b7-81dd278c0927",
		Name:         "Grim Flowering",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{N: b11CreatureCardsInGraveyard(ctx.Game, ctx.Controller())}.Apply(ctx)
		},
	})
}
