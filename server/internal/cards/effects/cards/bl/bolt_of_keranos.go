package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bolt of Keranos — Sorcery {1}{R}{R} (Born of the Gods,
// #5559):
//
//	"Bolt of Keranos deals 3 damage to any target. Scry 1. (Look at the
//	 top card of your library. You may put that card on the bottom.)"
//
// Jaya's Firenado's shape over any target: the damage, then the scry
// (CR 701.22a). A target that has become illegal counters the whole
// spell, scry included (CR 608.2b).
//
// No simplification.
func init() {
	purpose := effects.ForTargets(effects.DamageToTarget(0, 3))
	purpose.Scry = 1
	effects.Register(effects.Spec{
		OracleID:     "e07f50fb-98b2-40b7-9610-b8e9438be394",
		Name:         "Bolt of Keranos",
		Completeness: effects.CompletenessFull,
		Purpose:      purpose,
		Targets:      effects.TargetAny(),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok {
				if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
