package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raise the Past — Sorcery {2}{W}{W} (EDHREC rank 2801):
//
//	"Return all creature cards with mana value 2 or less from your
//	 graveyard to the battlefield."
//
// The white weenie mass reanimation. Every qualifying creature card
// in the caster's graveyard comes back under its owner's control,
// all of them as one entry (#1867), so its own enters-tapped clause
// runs and every ETB trigger sees the whole batch (CR 603.6a). Not targeted, as
// printed, and castable with an empty graveyard to no effect. Mana
// value is the printed cost's; a card with no cost is 0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a69a24d0-ca58-4a44-8af1-a3bd1608d2f9",
		Name:         "Raise the Past",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B26ReturnCreatureCardsWithManaValueAtMostFromGraveyard(ctx, ctx.Controller(), 2)
		},
	})
}
