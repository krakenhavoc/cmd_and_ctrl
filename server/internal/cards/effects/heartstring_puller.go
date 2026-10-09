package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heartstring Puller — Creature — Elf Sorcerer {3}{R}, 3/1:
//
//	"Trample
//	 When this creature enters, create a 2/2 colorless Wizard Soldier
//	 creature token named Cadet."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2cdcf65c-03d2-416f-8a3f-0322f75e8595",
		Name:            "Heartstring Puller",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Heartstring Puller — create a 2/2 Wizard Soldier named Cadet",
				Do(CreateToken{Template: TokenCard("2/2 colorless Wizard Soldier named Cadet"), N: 1})),
		},
	})
}
