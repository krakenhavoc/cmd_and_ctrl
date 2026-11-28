package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pollen Remedy — Instant {W} (Planeshift, S82 OLD b641, #5615):
//
//	"Kicker—Sacrifice a land. (You may sacrifice a land in addition to any other
//	 costs as you cast this spell.)
//	 Prevent the next 3 damage that would be dealt this turn to any number of
//	 targets, divided as you choose. If this spell was kicked, prevent the next 6
//	 damage this way instead."
//
// The division is announced with the targets (CR 601.2d), so the kicker carries
// its own clause with a total of 6: a kicked cast divides 6, an unkicked one 3.
// Each share is a shield that lasts the turn (CR 615.7); a target that is
// illegal on resolution takes no share and its share is not reassigned.
//
// No simplification.
func init() {
	kicker := effects.KickerSacrifice("a land", effects.Land())
	kicker.Targets = effects.TargetAny().WithCount(0, 0).Dividing(effects.Divide(6))
	effects.Register(effects.Spec{
		OracleID:      "ff52f955-98de-4f12-9ec7-f1d73e3edae9",
		Name:          "Pollen Remedy",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetAny().WithCount(0, 0).Dividing(effects.Divide(3)),
		OptionalCosts: []game.AdditionalCost{kicker},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.RprB04PreventDividedDamage("Pollen Remedy — prevent the next damage this turn")(ctx.Game, item)
		},
	})
}
