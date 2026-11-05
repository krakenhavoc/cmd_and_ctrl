package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Augur of Autumn — Creature — Human Druid {1}{G}{G}, 2/3:
//
//	"You may look at the top card of your library any time.
//	 You may play lands from the top of your library.
//	 Coven — As long as you control three or more creatures with
//	 different powers, you may cast creature spells from the top of
//	 your library."
//
// The look is private (LibraryTopOwner). The land permission is
// standing; the creature permission is gated on coven (trekB48Coven),
// read live each time it is asked, so it opens and closes as the board
// changes. Ordinary timing applies to both (the rulings).
//
// No simplification. Star Trek Commander (#4974).
func init() {
	Register(Spec{
		OracleID:          "9e774188-caa1-4b62-b9cc-547f8e444b7f",
		Name:              "Augur of Autumn",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(
				game.PermissionFilter{LandsOnly: true},
				"Play a land from the top of your library (Augur of Autumn)"),
		},
		GatedCastPermissions: []game.CastPermissionGate{{
			Permission: PlayFromTopOfYourLibrary(
				game.PermissionFilter{CreatureOnly: true},
				"Coven — cast a creature spell from the top of your library (Augur of Autumn)"),
			Condition: trekB48Coven,
		}},
	})
}
