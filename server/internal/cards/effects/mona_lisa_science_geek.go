package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mona Lisa, Science Geek — Legendary Creature — Lizard Mutant {2}{G}, 1/3:
//
//	"Reach
//	 {T}: Add X mana of any one color, where X is Mona Lisa's power."
//
// Reach is a printed keyword the deck importer stamps. The mana
// ability is #742's one pick of N tokens, with N read at activation
// from Mona Lisa's current power, so a pump counts and a -X/-0 that
// takes her to zero or below adds nothing (she still taps). Summoning
// sickness applies to the {T} cost as it does to any creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e6490e98-37a9-4a8b-a04c-cceeb22b0c35",
		Name:         "Mona Lisa, Science Geek",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{Tap: true},
			ProducedFunc: ProducedOneColor(func(g *game.Game, _, source uuid.UUID) int {
				c, ok := g.LookupCardForEffect(source)
				if !ok {
					return 0
				}
				return c.CurrentPower()
			}),
			Label: "Add X mana of any one color, where X is Mona Lisa's power",
		}},
	})
}
