package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voltage Surge — Instant {R} (Kamigawa: Neon Dynasty, S82 #5493):
//
//	"As an additional cost to cast this spell, you may sacrifice an
//	 artifact.
//	 Voltage Surge deals 2 damage to target creature or planeswalker. If
//	 this spell's additional cost was paid, Voltage Surge deals 4 damage
//	 instead."
//
// A plain optional additional cost (not a kicker), paid on casting
// (CR 601.2b, 601.2h); read back as "if the additional cost was paid".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1779af8e-38fc-4043-b5fa-ee16a7ea840c",
		Name:         "Voltage Surge",
		Completeness: effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{{
			Optional:  true,
			Key:       effects.OptionalAdditionalCostKey,
			Sacrifice: effects.SacrificeSpec("an artifact", effects.Artifact()),
			Label:     "Sacrifice an artifact",
		}},
		Targets: effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			amount := 2
			if ctx.OptionalCostTimes(effects.OptionalAdditionalCostKey) > 0 {
				amount = 4
			}
			return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: amount}.Apply(ctx)
		},
	})
}
