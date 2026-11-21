package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "4870ec65-8067-45fb-b7ec-a6eca87c7e77",
		Name:         "Ire of Kaminari",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
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
				return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
			}
			return nil
		},
	})
}
