package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sandstone Bridge — Land (Jumpstart 2022, #5948):
//
//	"This land enters tapped.
//	 When this land enters, target creature gets +1/+1 and gains
//	 vigilance until end of turn.
//	 {T}: Add {W}."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "08911e8e-cd67-4960-a927-958c33632469",
		Name:         "Sandstone Bridge",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Sandstone Bridge — target creature gets +1/+1 and gains vigilance until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok {
						return nil
					}
					if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1, Label: "Sandstone Bridge — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"vigilance"}, Label: "Sandstone Bridge — vigilance"}.Apply(ctx)
				}), effects.TargetCreature("target creature")),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
	})
}
