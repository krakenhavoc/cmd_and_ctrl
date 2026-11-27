package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Natural Affinity — Instant {2}{G} (Eighth Edition, S82 older sets batch 618, #5607):
//
//	"All lands become 2/2 creatures until end of turn. They're still lands."
//
// Rude Awakening's second mode over every land instead of yours: the set is
// the lands on the battlefield as this resolves and it is locked then
// (CR 611.2c), so a land that enters later is not animated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09a1ab1b-d9f0-4a2a-a448-3190f85006e4",
		Name:         "Natural Affinity",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ScopedEffectFor{
				Match: effects.Land(),
				Mods: []game.Mod{
					game.AddTypesMod("Creature"),
					game.SetBasePowerMod(2),
					game.SetBaseToughnessMod(2),
				},
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Natural Affinity — lands are 2/2 creatures",
			}.Apply(ctx)
		},
	})
}
