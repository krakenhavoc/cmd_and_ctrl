package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leafkin Avenger — Creature — Elemental Druid {2}{R}{G}, 4/3 (Core Set 2021,
// #5511):
//
//	"{T}: Add {G} for each creature with power 4 or greater you control.
//	 {7}{R}: This creature deals damage equal to its power to target player or
//	 planeswalker."
//
// The mana ability counts the creatures as it resolves, through the layered
// view (the Avenger itself counts at power 4 or more); a count of zero adds
// nothing and still taps. The damage is the Avenger's own, equal to its power
// as the ability resolves, and is not dealt for a power of 0 or less.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "adff8c13-8909-47d0-acad-ba63597f2e96",
		Name:         "Leafkin Avenger",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:           effects.ManaAbilityCost{Tap: true},
			ProducedAmount: effects.ProducedPerPermanent("G", func(c game.Card) bool { return c.IsCreature() && c.CurrentPower() >= 4 }),
			Label:          "Add {G} for each creature with power 4 or greater you control",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}{R}: This creature deals damage equal to its power to target player or planeswalker.",
			Cost:    effects.ManaCost("{7}{R}"),
			Targets: effects.TargetPlayerOrPlaneswalker(),
			Effect:  effects.SacrificedSourceDealsItsPowerToTheTarget,
		}},
	})
}
