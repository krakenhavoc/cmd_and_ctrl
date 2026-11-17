package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nekrataal — Creature — Human Assassin {2}{B}{B}, 2/1 (Visions, #5631):
//
//	"First strike
//	 When this creature enters, destroy target nonartifact, nonblack
//	 creature. That creature can't be regenerated."
//
// First strike is an engine keyword carried as printed data
// (CR 702.7b). The trigger targets as it goes on the stack (CR 603.3d);
// the rider is CR 701.19c, so a regeneration shield does not save it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d40adf5d-5edf-48a0-98f3-fcdbf1b5f25c",
		Name:            "Nekrataal",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Nekrataal — destroy target nonartifact, nonblack creature. That creature can't be regenerated",
					func(g *game.Game, item *game.StackItem) error {
						return destroyTheTargetPermanentNoRegen(item, NewContext(g, item))
					}),
				TargetCreature("target nonartifact, nonblack creature", Not(Artifact()), NonBlack())),
		},
	})
}
