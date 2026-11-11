package effects

import (
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
	Register(Spec{
		OracleID:     "b23cba46-6ed2-442d-bc39-7c4933f6b083",
		Name:         "Unholy Heat",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		Purpose:      ForTargets(DamageToTarget(0, 2)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			amount := 2
			if drcDeliriumMet(ctx.Game, ctx.Controller()) {
				amount = 6
			}
			return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: amount}.Apply(ctx)
		},
	})
}
