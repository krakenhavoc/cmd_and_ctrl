package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ire of Kaminari — Instant — Arcane {3}{R} (Betrayers of Kamigawa,
// #5602):
//
//	"Ire of Kaminari deals damage to any target equal to the number of
//	 Arcane cards in your graveyard."
//
// The count is taken as the spell resolves (CR 608.2h), while Ire of
// Kaminari is still on the stack, so it does not count itself. "Arcane"
// reads through a text change.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4870ec65-8067-45fb-b7ec-a6eca87c7e77",
		Name:         "Ire of Kaminari",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			arcane := ctx.TextSubtype("Arcane")
			n := 0
			if p := ctx.Game.PlayerByIDForEffect(ctx.Controller()); p != nil && p.Graveyard != nil {
				for _, c := range p.Graveyard.Cards {
					if c.HasSubtypeWord(arcane) {
						n++
					}
				}
			}
			for _, t := range ctx.LegalTargets() {
				return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
			}
			return nil
		},
	})
}
