package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fatal Push — Instant {B}:
//
//	"Destroy target creature if it has mana value 2 or less.
//	 Revolt — Destroy that creature if it has mana value 4 or less
//	 instead if a permanent left the battlefield under your control this
//	 turn."
//
// Any creature is a legal target; the mana value is read as the spell
// resolves (CR 608.2c), with ManaValueForEffect, so a card whose cost
// the engine cannot read is never destroyed rather than destroyed as if
// it were free. Revolt (revolt.go) is read at the same moment.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "16437a83-be52-44cd-a768-a767c9347eb2",
		Name:         "Fatal Push",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					continue
				}
				mv, known := ctx.Game.ManaValueForEffect(c)
				limit := 2
				if Revolt(ctx.Game, item.Controller) {
					limit = 4
				}
				if known && mv <= limit {
					return DestroyTarget{Target: t.ID}.Apply(ctx)
				}
			}
			return nil
		},
	})
}
