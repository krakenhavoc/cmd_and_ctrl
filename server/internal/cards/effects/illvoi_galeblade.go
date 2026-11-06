package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Illvoi Galeblade — Creature — Jellyfish Warrior {U}, 1/1 (Edge of
// Eternities, #4461):
//
//	"Flash
//	 Flying
//	 {2}, Sacrifice this creature: Draw a card."
//
// Yavimaya Elder's sacrifice-to-draw row on a flash flier. The sacrifice
// is part of the cost, so the card is drawn even if the Galeblade can't
// be found when the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d6273621-c36c-491b-85ca-37445ed30097",
		Name:            "Illvoi Galeblade",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice this creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
			Effect:  ebDrawACard,
		}},
	})
}
