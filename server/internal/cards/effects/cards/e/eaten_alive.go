package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eaten Alive — Sorcery {B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature or
//	 pay {3}{B}. Exile target creature or planeswalker."
//
// The either/or cost (ADR 0100 §2); the {3}{B} branch joins the total at
// CR 601.2f through the one pricer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d437ecc2-2fd3-4ad3-b23e-217f55e58dae",
		Name:         "Eaten Alive",
		Completeness: effects.CompletenessFull,
		AdditionalCost: effects.EitherCost(
			effects.SacrificeCost("a creature", effects.Creature()).Keyed("sacrifice"),
			effects.ManaAdditionalCost("{3}{B}").Keyed("mana"),
		),
		Targets: effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
