package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// First Volley — Instant — Arcane {1}{R} (Betrayers of Kamigawa, #5602):
//
//	"First Volley deals 1 damage to target creature and 1 damage to that
//	 creature's controller."
//
// One target: the controller is the creature's controller as the spell
// resolves (CR 608.2h). If the creature has become an illegal target the
// spell does nothing, so its controller isn't dealt damage either (CR
// 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b6a89457-1581-46d4-b75b-c5b8704d9f6f",
		Name:         "First Volley",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			controller := c.Controller
			if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: 1}).Apply(ctx); err != nil {
				return err
			}
			return DealDamage{Source: ctx.Source(), Target: controller, Amount: 1}.Apply(ctx)
		},
	})
}
