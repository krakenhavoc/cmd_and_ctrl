package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Piranha Marsh — Land (Jumpstart 2022, #5948):
//
//	"This land enters tapped.
//	 When this land enters, target player loses 1 life.
//	 {T}: Add {B}."
//
// Pierce Strider's targeted life loss on a tapped land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6bc8fd7e-7616-484a-ac23-04d37c93733b",
		Name:         "Piranha Marsh",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Piranha Marsh — target player loses 1 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return effects.GainLife{Player: t.ID, Amount: -1}.Apply(ctx)
					}
					return nil
				}), effects.TargetPlayer("target player")),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
	})
}
