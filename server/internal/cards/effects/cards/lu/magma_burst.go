package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magma Burst — Instant {3}{R} (Planeshift, S82 OLD b641, #5615):
//
//	"Kicker—Sacrifice two lands. (You may sacrifice two lands in addition to any
//	 other costs as you cast this spell.)
//	 Magma Burst deals 3 damage to any target. If this spell was kicked, it
//	 deals 3 damage to another target."
//
// Goblin Barrage's shape: the kicker's own target clause carries the second
// target, a different object from the first (CR 702.33d), and each is judged for
// itself on resolution (CR 608.2b).
//
// No simplification.
func init() {
	kicker := effects.OldB641KickerSacrificeN(2, "two lands", effects.Land())
	kicker.Targets = effects.Clauses(effects.TargetAny(), effects.Distinct(effects.TargetAny()))
	effects.Register(effects.Spec{
		OracleID:      "f5af6d60-decd-4e04-8841-da2d3000a405",
		Name:          "Magma Burst",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetAny(),
		Purpose:       effects.ForTargets(effects.DamageToTarget(0, 3)),
		OptionalCosts: []game.AdditionalCost{kicker},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.OldB208DamageClauseTarget(ctx, 0, 3) }),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return effects.OldB208DamageClauseTarget(ctx, 1, 3)
			}),
		),
	})
}
