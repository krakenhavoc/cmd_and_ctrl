package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gruesome Fate — Sorcery {2}{B} (Game Night):
//
//	"Each opponent loses 1 life for each creature you control."
//
// The creatures are counted once, as the spell resolves, and every opponent
// loses that much life (life loss, not damage).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf874feb-00fe-4e48-b57a-cf8b9855c057",
		Name:         "Gruesome Fate",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := effects.CountControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsCreature() })
			return effects.EachOpponentLosesLifeNow(ctx.Game, item, n)
		},
	})
}
