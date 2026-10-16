package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crag Puca — Creature — Shapeshifter {U/R}{U/R}{U/R}, 2/4:
//
//	"{U/R}: Switch this creature's power and toughness until end of turn."
//
// The switch is layer 7d (CR 613.4d, #3996): it applies after every
// set, pump and counter, whatever their timestamps.
// The ability's source must still be the same object (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b51d2c70-204b-4321-a589-f00a3e49fab4",
		Name:         "Crag Puca",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{U/R}: Switch this creature's power and toughness until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{U/R}"),
				Effect:  switchThisPT("Crag Puca — switch its power and toughness"),
			},
		},
	})
}
