package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// One with the Multiverse — Enchantment {6}{U}{U} (Duskmourn Commander):
//
//	"You may look at the top card of your library any time.
//	 You may play lands and cast spells from the top of your library.
//	 Once during each of your turns, you may cast a spell from your hand or
//	 the top of your library without paying its mana cost."
//
// Courser of Kruphix's and Future Sight's window (LibraryTopVisible and an
// unfiltered PlayFromTopOfYourLibrary: lands are played, spells cast at their
// mana cost), plus one granted free alternative cost (CR 118.9) reaching the
// hand and the top of the library, once during each of your turns per Multiverse
// (cmdr2S43CastFromHandOrTopFreeOnceEachOfYourTurns). A spell cast for free
// still has its own timing, and an {X} in its cost is 0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "06bb4272-25d1-46db-8585-15f919267e1e",
		Name:              "One with the Multiverse",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(game.PermissionFilter{}, "Play lands and cast spells from the top of your library (One with the Multiverse)"),
		},
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{
			cmdr2S43CastFromHandOrTopFreeOnceEachOfYourTurns(),
		},
	})
}
