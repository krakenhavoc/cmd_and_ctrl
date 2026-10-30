package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spider-Rex, Daring Dino — Legendary Creature — Spider Dinosaur Hero
// {4}{G}{G}, 6/6 (Marvel Super Heroes Commander, #4455):
//
//	"Reach, trample
//	 Ward {2}"
//
// Keywords only; ward is the engine's ward trigger (Blue Marvel,
// Adam Brashear declares it the same way).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "58b771d4-dfaf-4189-a5f6-4bc0fb4c0fa3",
		Name:            "Spider-Rex, Daring Dino",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Spider-Rex, Daring Dino — ward {2}"),
		},
	})
}
