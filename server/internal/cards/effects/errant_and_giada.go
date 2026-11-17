package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Errant and Giada — Legendary Creature — Human Angel {1}{W}{U}, 2/3:
//
//	"Flash
//	 Flying
//	 You may look at the top card of your library any time.
//	 You may cast spells with flash or flying from the top of your library."
//
// The Future Sight family (CR 401.5): a standing permission over the top
// card, and the private look that makes it usable. The keyword filter is
// PermissionFilter.KeywordsAny, read through the layered keyword view of
// the card on top, so a spell with either keyword qualifies and a land
// never does (a land is played, not cast, and has neither keyword).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "24697554-02fa-48ba-b51d-6bba3f85c293",
		Name:              "Errant and Giada",
		Completeness:      CompletenessFull,
		PrintedKeywords:   []string{"flash", "flying"},
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(
				game.PermissionFilter{NonLandOnly: true, KeywordsAny: "flash flying"},
				"Cast a spell with flash or flying from the top of your library (Errant and Giada)"),
		},
	})
}
