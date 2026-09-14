package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raise the Past — Sorcery {2}{W}{W} (EDHREC rank 2801):
//
//	"Return all creature cards with mana value 2 or less from your
//	 graveyard to the battlefield."
//
// The white weenie mass reanimation. Every qualifying creature card
// in the caster's graveyard comes back under its owner's control,
// each through the ordinary reanimation path, so its own
// enters-tapped clause and every ETB trigger fire. Not targeted, as
// printed, and castable with an empty graveyard to no effect. Mana
// value is the printed cost's; a card with no cost is 0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a69a24d0-ca58-4a44-8af1-a3bd1608d2f9",
		Name:         "Raise the Past",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b26ReturnCreatureCardsWithManaValueAtMostFromGraveyard(ctx, ctx.Controller(), 2)
		},
	})
}
