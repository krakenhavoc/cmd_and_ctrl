package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Garruk's Horde — Creature — Beast {5}{G}{G}, 7/7:
//
//	"Trample
//	 Play with the top card of your library revealed.
//	 You may cast creature spells from the top of your library. (Do
//	 this only any time you could cast that creature spell. You still
//	 pay the spell's costs.)"
//
// Trample rides PrintedKeywords. The revealed top and the creature-only
// permission are the Elven Chorus / Courser shapes: CreatureOnly keeps
// a land or a noncreature spell on top visible and unplayable, and the
// printed cost is paid. The reminder text is the engine's ordinary
// timing check, since a creature without flash is cast at sorcery speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "b9da511e-a57f-43c2-aa97-902f3b4c55fb",
		Name:              "Garruk's Horde",
		Completeness:      CompletenessFull,
		PrintedKeywords:   []string{"trample"},
		LibraryTopVisible: game.LibraryTopRevealed,
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(
				game.PermissionFilter{CreatureOnly: true},
				"Cast a creature spell from the top of your library (Garruk's Horde)"),
		},
	})
}
