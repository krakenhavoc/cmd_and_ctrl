package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archive Dragon — Creature — Dragon Wizard {4}{U}{U}, 4/6 (Wilds of
// Eldraine, #4471):
//
//	"Flying
//	 Ward {2} (Whenever this creature becomes the target of a spell or
//	 ability an opponent controls, counter it unless that player pays
//	 {2}.)
//	 When this creature enters, scry 2."
//
// Benevolent River Spirit without the waterbend: flying is a printed
// keyword, ward {2} is the catalog's ward trigger, and the entry scry is
// the Scry primitive on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cddbd819-4895-4712-b44b-d6b51f3d8646",
		Name:            "Archive Dragon",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Scry: 2},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Archive Dragon — ward {2}"),
			WhenThisEnters("Archive Dragon — scry 2", Do(Scry{N: 2})),
		},
	})
}
