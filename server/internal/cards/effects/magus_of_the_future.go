package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magus of the Future — Creature — Human Wizard {2}{U}{U}{U}, 2/3:
//
//	"Play with the top card of your library revealed.
//	 You may play lands and cast spells from the top of your library."
//
// Future Sight on a body: the REVEALED visibility strength (every seat
// sees the top card) and an unfiltered library permission, so lands are
// played and spells cast off the top at their printed cost. Both halves
// are derived from the battlefield and end when the Magus leaves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "10758aa7-0f64-499f-8010-514e8fc31b7e",
		Name:              "Magus of the Future",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopRevealed,
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(game.PermissionFilter{},
				"Play a land or cast a spell from the top of your library (Magus of the Future)"),
		},
	})
}
