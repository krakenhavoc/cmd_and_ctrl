package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shinka, the Bloodsoaked Keep — Legendary Land:
//
//	"{T}: Add {R}.
//	 {R}, {T}: Target legendary creature gains first strike until end
//	 of turn."
//
// Shizo, Death's Storehouse's shape with {R} and first strike. The
// target is any legendary creature, not only yours.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "255b937f-c9c9-4ae9-815e-17418eba0602",
		Name:         "Shinka, the Bloodsoaked Keep",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, {T}: Target legendary creature gains first strike until end of turn",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "first strike")),
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost()),
			Targets: effects.TargetCreature("target legendary creature", effects.Legendary()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, target := range ctx.LegalTargets() {
					if target.Kind != game.TargetCard {
						continue
					}
					return (effects.GrantKeywordUntilEOT{
						Target:   target.ID,
						Keywords: []string{"first strike"},
						Label:    "Shinka, the Bloodsoaked Keep — first strike",
					}).Apply(ctx)
				}
				return nil
			},
		}},
	})
}
