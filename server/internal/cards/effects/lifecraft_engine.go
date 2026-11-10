package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lifecraft Engine — Artifact — Vehicle {3}, 4/4 (Aetherdrift, #4464):
//
//	"As this Vehicle enters, choose a creature type.
//	 Vehicle creatures you control are the chosen creature type in addition
//	 to their other types.
//	 Each creature you control of the chosen type other than this Vehicle
//	 gets +1/+1.
//	 Crew 3"
//
// The choice is Adaptive Automaton's CR 614.12 entry prompt, stored on
// Card.NamedTribe. The type-add is a layer 4 static over Vehicles you
// control that are creatures (it applies to a crewed Vehicle, the Engine
// itself included, and to no uncrewed one; crew's layer 4 effect is applied
// first by CR 613.8 dependency). The anthem is a layer 7c static over
// creatures of the chosen type you control other than the Engine, so it
// reads the type the first static added: a crewed Vehicle of the chosen type
// gets +1/+1 too. Until a type is named both do nothing. Crew is the
// engine's (CR 702.122a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "caa69b51-1efd-4d0e-803b-914a83ccb836",
		Name:         "Lifecraft Engine",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Lifecraft Engine"),
		Static: []game.StaticAbility{
			vehicleCreaturesAreAlsoTheChosenType(),
			TribalAnthem(TribeFilter{Chosen: true, Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Lifecraft Engine"),
		}},
	})
}
