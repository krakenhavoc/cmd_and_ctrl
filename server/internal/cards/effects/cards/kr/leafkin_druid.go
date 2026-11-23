package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Leafkin Druid — Creature — Elemental Druid {1}{G}, 0/3:
//
//	"{T}: Add {G}. If you control four or more creatures, add {G}{G}
//	 instead."
//
// Raucous Audience's shape with a creature count: one mana ability whose
// output is decided as it resolves, by the number of creatures its
// controller controls (the Druid itself included).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "55eb9310-f24b-410b-850b-a7d0d5117946",
		Name:         "Leafkin Druid",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{Tap: true},
			ProducedFunc: func(g *game.Game, controller, _ uuid.UUID) string {
				n := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == controller && c.IsCreature() {
						n++
					}
				}
				if n >= 4 {
					return "{G}{G}"
				}
				return "{G}"
			},
			Label: "Add {G} ({G}{G} with four or more creatures)",
		}},
	})
}
