package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Piety — Instant {2}{W} (Arabian Nights, #5641):
//
//	"Blocking creatures get +0/+3 until end of turn."
//
// Army of Allah's mirror for the defence. The set of blocking creatures
// is locked in as the spell resolves, whoever controls them; a creature
// that blocks later in the turn gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c017406-7fc3-4701-93ec-ddb02044c12a",
		Name:         "Piety",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{
				Match:     effects.BlockingCreature(),
				Toughness: 3,
				Label:     "Piety — blocking creatures get +0/+3",
			}.Apply(ctx)
		},
	})
}
