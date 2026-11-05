package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Benjamin Sisko, Besieged — Legendary Creature — Human Officer
// {4}{G}{G}, 5/5:
//
//	"Play with the top card of your library revealed.
//	 You may play lands and cast creature spells from the top of your
//	 library."
//
// Courser of Kruphix's revealed top and land permission plus Garruk's
// Horde's creature permission: two standing permissions, one per kind
// of card, both on the same revealed top card (CR 401.5).
//
// No simplification. Star Trek Commander (#4974).
func init() {
	Register(Spec{
		OracleID:          "f1416577-97d0-4d6c-afd2-455798556d4c",
		Name:              "Benjamin Sisko, Besieged",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopRevealed,
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(
				game.PermissionFilter{LandsOnly: true},
				"Play a land from the top of your library (Benjamin Sisko, Besieged)"),
			PlayFromTopOfYourLibrary(
				game.PermissionFilter{CreatureOnly: true},
				"Cast a creature spell from the top of your library (Benjamin Sisko, Besieged)"),
		},
	})
}
