package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Centaur Nurturer — Creature — Centaur Druid {3}{G}, 2/4 (War of the Spark, #5520):
//
//	"When this creature enters, you gain 3 life.
//	 {T}: Add one mana of any color."
//
// Dawnhart Rejuvenator's two abilities.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f60e831b-1404-437a-a587-ed4f4b596084",
		Name:         "Centaur Nurturer",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Centaur Nurturer — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
}
