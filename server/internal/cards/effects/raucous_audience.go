package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Raucous Audience — Creature — Human Citizen {1}{G}, 2/1:
//
//	"{T}: Add {G}. If you control a creature with power 4 or greater,
//	 add {G}{G} instead."
//
// Ilysian Caryatid's shape in mono-green: one mana ability whose output
// is decided as it resolves, by whether its controller controls a
// creature with power 4 or greater (itself included, should it get
// there), read through the layered view.
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "5be6f245-31cf-4294-ace0-ca6ce3463fed",
		Name:         "Raucous Audience",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{Tap: true},
			ProducedFunc: func(g *game.Game, controller, _ uuid.UUID) string {
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == controller && c.IsCreature() && c.CurrentPower() >= 4 {
						return "{G}{G}"
					}
				}
				return "{G}"
			},
			Label: "Add {G} ({G}{G} with a 4-power creature)",
		}},
	})
}
