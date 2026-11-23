package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shizo, Death's Storehouse — "{T}: Add {B}. {B}, {T}: Target legendary
// creature gains fear until end of turn. (It can't be blocked except by
// artifact creatures and/or black creatures.)"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "008f2698-1721-45a3-8353-10f2f400dc8f",
		Name:         "Shizo, Death's Storehouse",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, {T}: Target legendary creature gains fear until end of turn",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "fear")),
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.TapCost()),
			Targets: effects.TargetCreature("target legendary creature", effects.Legendary()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, target := range ctx.LegalTargets() {
					if target.Kind != game.TargetCard {
						continue
					}
					return (effects.GrantKeywordUntilEOT{
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
