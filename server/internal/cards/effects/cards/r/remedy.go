package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Remedy — Instant {1}{W} (Dominaria Remastered, #5948):
//
//	"Prevent the next 5 damage that would be dealt this turn to any number
//	 of targets, divided as you choose."
//
// Angel of Salvation's clause as a spell: each target still legal gets its
// own shield for the share announced for it (CR 601.2d, 615.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c644f190-63e5-4496-8612-bbe7be91afc9",
		Name:         "Remedy",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny().WithCount(0, 0).Dividing(effects.Divide(5)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.RprB04PreventDividedDamage("Remedy — prevent the next damage this turn")(ctx.Game, item)
		},
	})
}
