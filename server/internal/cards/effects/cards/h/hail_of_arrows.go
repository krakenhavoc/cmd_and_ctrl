package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hail of Arrows — Instant {X}{W} (Conspiracy: Take the Crown, S82 older
// sets batch 270, #5540):
//
//	"Hail of Arrows deals X damage divided as you choose among any number
//	 of target attacking creatures."
//
// Rolling Thunder's shape (any number of targets, X divided, CR 601.2d)
// over the attacking creatures: every pick must be an attacking creature
// as the spell is cast and again as it resolves, and a target that has
// stopped attacking takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "37429f88-e29b-4b1b-90fd-23a04d703c0a",
		Name:         "Hail of Arrows",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetCreature("any number of target attacking creatures", effects.AttackingCreature()).WithCount(0, 0).Dividing(effects.DivideX()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
}
