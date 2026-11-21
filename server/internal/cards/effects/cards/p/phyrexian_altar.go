package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Altar — Artifact {3}:
//
//	"Sacrifice a creature: Add one mana of any color."
//
// Ashnod's Altar's coloured cousin: one mana instead of two, but any
// colour, which is what makes it the combo piece of choice when the
// loop needs a specific colour rather than raw quantity.
//
// The five-colour pipe offers all five colours, the controller's
// commander identity listed first, as Treasure and Birds of Paradise
// do.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8d02b297-97c4-4379-9862-0a462400f66f",
		Name:         "Phyrexian Altar",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{
				SacrificeOther: effects.SacrificeACreature().SacrificeOther,
			},
			Produced: "{W|U|B|R|G}",
			Label:    "Sacrifice a creature: Add one mana of any color",
			Answers:  game.AnswerSacOutlet,
		}},
	})
}
