package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hurloon Battle Hymn — {2}{R} Instant (Dominaria United):
//
//	"Kicker {W} (You may pay an additional {W} as you cast this spell.)
//	 Hurloon Battle Hymn deals 4 damage to target creature or
//	 planeswalker. If this spell was kicked, you gain 4 life."
//
// The life is gained even when the damage kills nothing, but not when the
// only target is gone: the spell then does not resolve (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:      "2f541bf9-1e5b-4d16-8558-a5cfce7e93ad",
		Name:          "Hurloon Battle Hymn",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 4)), Answers: game.AnswerRemove, Kicked: &game.Purpose{LifeGain: 4}},
		Targets:       TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OptionalCosts: []game.AdditionalCost{Kicker("{W}")},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				return DealDamage{Target: t.ID, Amount: 4, Source: ctx.Item.SourceCardID}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return GainLife{Amount: 4}.Apply(ctx)
			}),
		),
	})
}
