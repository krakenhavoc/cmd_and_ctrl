package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enlightened Maniac — Creature — Human {3}{U}, 0/2 (Eldritch Moon, #5541):
//
//	"When this creature enters, create a 3/2 colorless Eldrazi Horror
//	 creature token."
//
// The 3/2 Eldrazi Horror is Howling Chorus's token row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4310524d-dc0e-48ad-ae66-4c64aaa39340",
		Name:         "Enlightened Maniac",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Enlightened Maniac — create a 3/2 colorless Eldrazi Horror creature token",
				Do(CreateToken{Template: TokenCard("3/2 colorless Eldrazi Horror"), N: 1})),
		},
	})
}
