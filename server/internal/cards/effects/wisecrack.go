package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wisecrack — Instant {2}{R} (Marvel Super Heroes Commander, #4455):
//
//	"Target creature deals damage equal to its power to itself. If that
//	 creature is attacking, Wisecrack deals 2 damage to that creature's
//	 controller."
//
// The creature is the damage source (so its deathtouch or lifelink apply),
// read for power as the spell resolves. The attacking check follows the
// damage on the same object; the 2 damage is Wisecrack's, dealt to the
// creature's controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e9c9a445-4386-4616-ad00-f08064d73dcd",
		Name:         "Wisecrack",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			t := targets[0]
			c, ok := ctx.Game.LookupCardForEffect(t.ID)
			if !ok {
				return nil
			}
			attacking, controller := c.AttackingTarget != uuid.Nil, c.Controller
			if err := (DealDamage{Source: t.ID, Target: t.ID, Amount: c.CurrentPower()}).Apply(ctx); err != nil {
				return err
			}
			if !attacking {
				return nil
			}
			return DealDamage{Source: item.SourceCardID, Target: controller, Amount: 2}.Apply(ctx)
		},
	})
}
