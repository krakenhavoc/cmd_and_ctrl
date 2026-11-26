package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spawn of Thraxes — Creature — Dragon {5}{R}{R}, 5/5 (Journey into Nyx, S82
// OLD, #5558):
//
//	"Flying
//	 When this creature enters, it deals damage to any target equal to the
//	 number of Mountains you control."
//
// The Mountains are counted as the trigger resolves, and the Spawn is the
// damage source, so lifelink or protection it has is respected.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "52289774-d292-4028-91fb-8826fa004dc8",
		Name:            "Spawn of Thraxes",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Spawn of Thraxes — it deals damage to any target equal to the number of Mountains you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						n := effects.CountControlled(g, ctx.Controller(), func(p game.Card) bool { return p.HasSubtypeWord("Mountain") })
						return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
					}
					return nil
				}),
				effects.TargetAny()),
		},
	})
}
