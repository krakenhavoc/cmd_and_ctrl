package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Veloheart Bike — Artifact — Vehicle {2}{G}, 4/2 (Aetherdrift, #4464):
//
//	"When this Vehicle enters, you gain 2 life.
//	 {T}: Add one mana of any color.
//	 Crew 2 (Tap any number of creatures you control with total power 2
//	 or more: This Vehicle becomes an artifact creature until end of
//	 turn.)"
//
// The enters trigger uses the stack. The tap is a mana ability
// (CR 605.1a): an uncrewed Vehicle is not a creature, so summoning
// sickness does not stop it, while a crewed one that came under its
// controller's control this turn cannot tap for mana (CR 302.6). Crew is
// the engine's (CR 702.122a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa123b21-518a-4f1d-9629-c44576d45f21",
		Name:         "Veloheart Bike",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Veloheart Bike — you gain 2 life", Do(GainLife{Amount: 2})),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color.",
		}},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Veloheart Bike"),
		}},
	})
}
