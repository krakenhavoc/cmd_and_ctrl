package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shizo, Death's Storehouse — "{T}: Add {B}. {B}, {T}: Target legendary
// creature gains fear until end of turn. (It can't be blocked except by
// artifact creatures and/or black creatures.)"
func init() {
	Register(Spec{
		OracleID:     "008f2698-1721-45a3-8353-10f2f400dc8f",
		Name:         "Shizo, Death's Storehouse",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{B}, {T}: Target legendary creature gains fear until end of turn",
			Cost:    Plus(ManaCost("{B}"), TapCost()),
			Targets: TargetCreature("target legendary creature", Legendary()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, target := range ctx.LegalTargets() {
					if target.Kind != game.TargetCard {
						continue
					}
					return (GrantKeywordUntilEOT{
						Target:   target.ID,
						Keywords: []string{"fear"},
						Label:    "Shizo, Death's Storehouse — fear",
					}).Apply(ctx)
				}
				return nil
			},
		}},
	})
}
