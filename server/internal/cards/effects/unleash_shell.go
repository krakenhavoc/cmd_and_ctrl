package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unleash Shell — Instant for {3}{R}{R} (The Brothers' War):
//
//	"Unleash Shell deals 5 damage to target creature or planeswalker and
//	2 damage to that permanent's controller."
//
// One sentence, one damage instance (CR 615.8). The controller is read off
// the permanent before the damage, so a lethal hit that destroys it does not
// lose the second half (CR 608.2h).
func init() {
	Register(Spec{
		OracleID:     "9eaca93d-6c3d-4489-8652-b575a4cd931f",
		Name:         "Unleash Shell",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		Purpose:      ForTargets(DamageToTarget(0, 5)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			victim, ok := ctx.Game.LookupCardForEffect(t.ID)
			if !ok {
				return nil
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 5}).Apply(ctx); err != nil {
					return err
				}
				return DealDamage{Source: ctx.Source(), Target: victim.Controller, Amount: 2}.Apply(ctx)
			})
		},
	})
}
