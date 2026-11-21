package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hurloon Battle Hymn — {2}{R} Instant (Dominaria United):
//
//	"Kicker {W} (You may pay an additional {W} as you cast this spell.)
//	 Hurloon Battle Hymn deals 4 damage to target creature or
//	 planeswalker. If this spell was kicked, you gain 4 life."
//
// The life is gained even when the damage kills nothing, but not when the
// only target is gone: the spell then does not resolve (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:      "2f541bf9-1e5b-4d16-8558-a5cfce7e93ad",
		Name:          "Hurloon Battle Hymn",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 4)), Answers: game.AnswerRemove, Kicked: &game.Purpose{LifeGain: 4}},
		Targets:       effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{W}")},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				return effects.DealDamage{Target: t.ID, Amount: 4, Source: ctx.Item.SourceCardID}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return effects.GainLife{Amount: 4}.Apply(ctx)
			}),
		),
	})
}
