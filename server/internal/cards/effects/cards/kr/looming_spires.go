package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Looming Spires — Land (Battle for Zendikar):
//
//	"This land enters tapped.
//	 When this land enters, target creature gets +1/+1 and gains first strike
//	 until end of turn.
//	 {T}: Add {R}."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d09b136-525f-49dd-a3a2-dfaca4e8e9a8",
		Name:         "Looming Spires",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Looming Spires — target creature gets +1/+1 and gains first strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := (effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Looming Spires — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike"}, Label: "Looming Spires — first strike"}.Apply(ctx)
				}), effects.TargetCreature("target creature")),
		},
	})
}
