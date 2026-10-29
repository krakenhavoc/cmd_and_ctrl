package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bitter Downfall — Instant {3}{B} (The Hobbit Commander):
//
//	"This spell costs {3} less to cast if it targets a creature that was
//	 dealt damage this turn.
//	 Destroy target creature. Its controller loses 2 life."
//
// #4454, S76. The reduction reads the announced target
// (CostsLessIfItTargets) with the shared DealtDamageThisTurn predicate.
// The creature's controller is read before the destroy, so the life loss
// still lands on them when the creature is regenerated or dies, and the
// whole spell does nothing when the target has left (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "19c22e08-29b6-4b3d-bf6e-28a5add620bb",
		Name:         "Bitter Downfall",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(3, "This spell costs {3} less to cast if it targets a creature that was dealt damage this turn.",
				And(Creature(), DealtDamageThisTurn())),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					continue
				}
				controller := c.Controller
				if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), controller, -2)
			}
			return nil
		},
	})
}
