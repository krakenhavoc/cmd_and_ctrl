package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sarkhan's Rage — Instant {4}{R} (Foundations Jumpstart):
//
//	"Sarkhan's Rage deals 5 damage to any target. If you control no
//	 Dragons, Sarkhan's Rage deals 2 damage to you."
//
// The 5 damage, then the self-damage if the caster controls no Dragon
// permanent as the second sentence is read (CR 608.2c), after the first
// damage and whatever it killed. The Dragon check reads the layered
// battlefield, so a changeling or a creature made a Dragon counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3313aa8f-6441-4daa-b8fa-e6144f60bea8",
		Name:         "Sarkhan's Rage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 5)), Answers: game.AnswerRemove},
		Targets:      effects.TargetAny(),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.S82B48DamageToTheFirstTarget(5)),
			effects.Instruction(func(ctx *effects.Context) error {
				if effects.B05ControlsSubtype(ctx.Game, ctx.Controller(), ctx.TextSubtype("Dragon")) {
					return nil
				}
				return effects.DealDamage{Source: ctx.Source(), Target: ctx.Controller(), Amount: 2}.Apply(ctx)
			}),
		),
	})
}
