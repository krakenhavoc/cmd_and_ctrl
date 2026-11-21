package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Future Sight — Enchantment {2}{U}{U}{U}:
//
//	"Play with the top card of your library revealed.
//	 You may play lands and cast spells from the top of your library."
//
// The Oracle of Mul Daya / Bolas's Citadel pair of slots: the top card is
// revealed to everyone (LibraryTopRevealed) and a permission with no filter
// lets its controller play a land (as their land drop) or cast a spell from
// the top (CR 401.5), paying the printed cost.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "93a0e938-b916-47ec-9803-369e885e253c",
		Name:              "Future Sight",
		Completeness:      effects.CompletenessFull,
		LibraryTopVisible: game.LibraryTopRevealed,
		CastPermissions: []game.CastPermission{
			effects.PlayFromTopOfYourLibrary(game.PermissionFilter{},
				"Play a land or cast a spell from the top of your library (Future Sight)"),
		},
	})
}
