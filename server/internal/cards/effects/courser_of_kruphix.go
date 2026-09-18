package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Courser of Kruphix — Enchantment Creature — Centaur {1}{G}{G}, 2/4:
//
//	"Play with the top card of your library revealed.
//	 You may play lands from the top of your library.
//	 Landfall — Whenever a land you control enters, you gain 1 life."
//
// Oracle of Mul Daya's two static halves without the extra land drop,
// plus the landfall that makes them a life engine: the land you play
// off the top triggers the third line, so the card gains life from its
// own permission. Nothing had to connect those two — the land play
// from the library goes through CastSpell's ordinary land branch and
// emits the same EventETB any other land entry does, which is what
// Landfall watches.
//
// The visibility is the REVEALED strength (CR 401.5), so every seat
// sees the top card of a Courser controller's library and the
// per-viewer filter passes exactly that one card through an otherwise
// wholesale-hidden zone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "46779609-4fa7-4fd2-b5b4-7d4d749339e6",
		Name:              "Courser of Kruphix",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopRevealed,
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(
				game.PermissionFilter{LandsOnly: true},
				"Play a land from the top of your library (Courser of Kruphix)"),
		},
		Triggered: []game.TriggeredAbility{
			Landfall("Courser of Kruphix — you gain 1 life (landfall)", Do(GainLife{Amount: 1})),
		},
	})
}
