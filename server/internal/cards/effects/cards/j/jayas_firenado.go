package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jaya's Firenado — Sorcery {4}{R} (Dominaria United, #5487):
//
//	"Jaya's Firenado deals 5 damage to target creature or planeswalker.
//	 Scry 1."
//
// The damage, then the scry (CR 701.22a). A target that has become illegal
// counters the whole spell, scry included (CR 608.2b).
//
// No simplification.
func init() {
	purpose := effects.ForTargets(effects.DamageToTarget(0, 5))
	purpose.Scry = 1
	effects.Register(effects.Spec{
		OracleID:     "250f804f-b5d5-414c-bb25-476d1fe1f85a",
		Name:         "Jaya's Firenado",
		Completeness: effects.CompletenessFull,
		Purpose:      purpose,
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 5}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
