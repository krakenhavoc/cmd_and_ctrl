package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unholy Heat — Instant {R} (Outlaws of Thunder Junction Commander):
//
//	"Unholy Heat deals 2 damage to target creature or planeswalker.
//	 Delirium — Unholy Heat deals 6 damage instead if there are four or more
//	 card types among cards in your graveyard."
//
// Delirium is an ability word (CR 207.2c). The amount is read as the spell
// resolves (the spell itself is on the stack, so it does not count), from the
// caster's own graveyard (drcDeliriumMet), and the damage goes to the target
// if it is still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b23cba46-6ed2-442d-bc39-7c4933f6b083",
		Name:         "Unholy Heat",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 2)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			amount := 2
			if effects.DrcDeliriumMet(ctx.Game, ctx.Controller()) {
				amount = 6
			}
			return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: amount}.Apply(ctx)
		},
	})
}
