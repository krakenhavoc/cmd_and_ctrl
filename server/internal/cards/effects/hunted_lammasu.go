package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hunted Lammasu — Creature — Lammasu {2}{W}{W}, 5/5 (Ravnica: City of
// Guilds, #5600):
//
//	"Flying
//	 When this creature enters, target opponent creates a 4/4 black Horror
//	 creature token."
//
// Hunted Horror's targeted enters trigger: the opponent creates the
// Horror, so they own and control it (CR 111.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a76667b5-2144-4395-a5c1-57cb3c29a99e",
		Name:            "Hunted Lammasu",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			ravB4TargetOpponentCreates("Hunted Lammasu — target opponent creates a 4/4 black Horror", "4/4 black Horror", 1),
		},
	})
}
