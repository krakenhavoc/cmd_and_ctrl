package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dire Fleet Ravager — Creature — Orc Pirate Wizard {3}{B}{B}, 4/4:
//
//	"Menace, deathtouch
//	 When this creature enters, each player loses a third of their life,
//	 rounded up."
//
// "Each player" includes the Ravager's controller. The amounts are read
// off every life total before any loss is applied, so one player's loss
// can't change another's (CR 608.2c), and a player at or below 0 loses
// nothing more.
func init() {
	Register(Spec{
		OracleID:        "b83080af-9616-4dc3-bc47-6ef84b9ccc1b",
		Name:            "Dire Fleet Ravager",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dire Fleet Ravager — each player loses a third of their life, rounded up",
				cmdr2B05EachPlayerLosesAThirdOfTheirLife),
		},
	})
}
