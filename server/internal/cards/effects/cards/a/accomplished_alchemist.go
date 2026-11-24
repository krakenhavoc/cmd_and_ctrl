package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Accomplished Alchemist — {3}{G} Creature — Elf Druid 2/5 (Strixhaven:
// School of Mages, #5502):
//
//	"{T}: Add one mana of any color.
//	 {T}: Add X mana of any one color, where X is the amount of life you
//	 gained this turn."
//
// Two mana abilities. The second is Mona Lisa's shape (N mana of one picked
// colour, counted as it is activated), with N the life its controller has
// gained this turn, the same tally Blossoming Bogbeast reads: zero gains
// produce nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97e18a8d-9068-4782-a744-280d28c68228",
		Name:         "Accomplished Alchemist",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
			},
			{
				Cost: effects.ManaAbilityCost{Tap: true},
				ProducedAmount: effects.ProducedOneColor(func(g *game.Game, controller, _ uuid.UUID) int {
					return effects.B15LifeGainedThisTurn(g, controller)
				}),
				Label: "Add X mana of any one color, where X is the amount of life you gained this turn",
			},
		},
	})
}
