package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Korlessa, Scale Singer — Legendary Creature — Dragon Bard {G}{U}, 1/4
// (Commander Legends: Battle for Baldur's Gate):
//
//	"You may look at the top card of your library any time.
//	 You may cast Dragon spells from the top of your library."
//
// Both halves are Conspicuous Snoop's seams. The look is LibraryTopOwner
// (only the controller sees the top card, CR 401.5, and sees it as it changes).
// The cast permission is a cast-only top-of-library permission (CR 601.2) for
// nonland spells with the Dragon creature type, at printed cost and timing; a
// changeling spell is a Dragon spell (CR 702.73a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "2ceac1be-32ab-4a04-ade3-852926a5ca45",
		Name:              "Korlessa, Scale Singer",
		Completeness:      effects.CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{
			effects.PlayFromTopOfYourLibrary(
				game.PermissionFilter{NonLandOnly: true, CreatureType: "Dragon"},
				"Cast a Dragon spell from the top of your library (Korlessa, Scale Singer)"),
		},
	})
}
