package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ilysian Caryatid — Creature — Plant {1}{G}, 1/1:
//
//	"{T}: Add one mana of any color. If you control a creature with
//	 power 4 or greater, add two mana of any one color instead."
//
// One ability whose output is decided at activation: the ordinary
// any-colour pick, or #742's one pick of two tokens when the
// controller controls a creature with power 4 or greater (the Caryatid
// itself included, should it ever get there). Two same-colour tokens,
// never one of each — that is what "any one color" means, and what a
// second independent pick would have got wrong.
//
// "Any color", flatly — nothing about the commander's identity.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f2be3c2-060a-43e1-b63b-9cd3c78ffcb0",
		Name:         "Ilysian Caryatid",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{Tap: true},
			ProducedFunc: func(g *game.Game, controller, _ uuid.UUID) string {
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == controller && c.IsCreature() && c.CurrentPower() >= 4 {
						return OneColorOfAmount(2)
					}
				}
				return "{W|U|B|R|G}"
			},
			Label:                   "Add one mana of any color (two of any one color with a 4-power creature)",
			IgnoreCommanderIdentity: true,
		}},
	})
}
