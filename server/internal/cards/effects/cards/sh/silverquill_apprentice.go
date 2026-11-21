package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silverquill Apprentice — Creature — Human Warlock {W}{B}, 2/2
// (Strixhaven: School of Mages, #5502):
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery spell,
//	 target creature gets +1/+0 until end of turn."
//
// The shared Magecraft trigger with a target (Octavia, Living Thesis's
// shape): any creature, chosen as the trigger goes on the stack, gets
// +1/+0 if it is still a legal target as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ab91144-5eb3-491b-a8be-7e1139d0d6c8",
		Name:         "Silverquill Apprentice",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Magecraft("Silverquill Apprentice — target creature gets +1/+0 until end of turn (magecraft)",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetCard {
								return effects.BoostUntilEOT{Target: t.ID, Power: 1, Label: "Silverquill Apprentice — +1/+0"}.Apply(ctx)
							}
						}
						return nil
					}),
				effects.TargetCreature("target creature")),
		},
	})
}
