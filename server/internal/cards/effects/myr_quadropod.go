package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Myr Quadropod — Artifact Creature — Myr {4}, 1/4:
//
//	"{3}: Switch this creature's power and toughness until end of turn."
//
// The switch is layer 7d (CR 613.4d, #3996): it applies after every
// set, pump and counter, whatever their timestamps.
// The ability's source must still be the same object (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f53732d3-d317-4042-a0ad-1a99c7ffc94f",
		Name:         "Myr Quadropod",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{3}: Switch this creature's power and toughness until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{3}"),
				Effect:  switchThisPT("Myr Quadropod — switch its power and toughness"),
			},
		},
	})
}
