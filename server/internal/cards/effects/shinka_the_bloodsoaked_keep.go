package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "255b937f-c9c9-4ae9-815e-17418eba0602",
		Name:         "Shinka, the Bloodsoaked Keep",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{R}, {T}: Target legendary creature gains first strike until end of turn",
			Cost:    Plus(ManaCost("{R}"), TapCost()),
			Targets: TargetCreature("target legendary creature", Legendary()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, target := range ctx.LegalTargets() {
					if target.Kind != game.TargetCard {
						continue
					}
					return (GrantKeywordUntilEOT{
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
