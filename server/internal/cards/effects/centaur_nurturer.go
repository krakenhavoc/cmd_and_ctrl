package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Centaur Nurturer — Creature — Centaur Druid {3}{G}, 2/4 (War of the Spark, #5520):
//
//	"When this creature enters, you gain 3 life.
//	 {T}: Add one mana of any color."
//
// Dawnhart Rejuvenator's two abilities.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f60e831b-1404-437a-a587-ed4f4b596084",
		Name:         "Centaur Nurturer",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color.",
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Centaur Nurturer — you gain 3 life", Do(GainLife{Amount: 3})),
		},
	})
}
