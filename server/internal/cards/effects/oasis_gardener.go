package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oasis Gardener — Artifact Creature — Scarecrow {3}, 2/2 (Outlaws of
// Thunder Junction, #4468):
//
//	"When this creature enters, you gain 2 life.
//	 {T}: Add one mana of any color."
//
// Alloy Myr's mana ability on a body with an enters life gain. The {T}
// is a creature's, so it can't be paid the turn the Gardener arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "210a130c-d5d3-43bf-8a80-d796c37242e8",
		Name:         "Oasis Gardener",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Oasis Gardener — you gain 2 life", Do(GainLife{Amount: 2})),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
