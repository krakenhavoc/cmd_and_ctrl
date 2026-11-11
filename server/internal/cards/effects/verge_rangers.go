package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Verge Rangers — Creature — Human Scout Ranger {2}{W}, 3/3 (Duskmourn
// Commander, #5368):
//
//	"First strike
//	 You may look at the top card of your library any time.
//	 As long as an opponent controls more lands than you, you may play lands from
//	 the top of your library. (You can play a land this way only if you have an
//	 available land play remaining.)"
//
// Augur of Autumn's shape: the look is LibraryTopVisible (the owner may look at
// the top card any time) and the land play is a gated library permission whose
// condition, "an opponent controls more lands than you", is read live by the
// permission's liveness test, so it is dormant while you are level or ahead and
// live again when an opponent pulls ahead. Playing a land this way uses the
// turn's land play (CR 305.2), as the reminder says.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "aece819e-3678-48bc-9677-d06ebfa58b4e",
		Name:              "Verge Rangers",
		Completeness:      CompletenessFull,
		PrintedKeywords:   []string{"first strike"},
		LibraryTopVisible: game.LibraryTopOwner,
		GatedCastPermissions: []game.CastPermissionGate{{
			Permission: PlayFromTopOfYourLibrary(
				game.PermissionFilter{LandsOnly: true},
				"Play a land from the top of your library (Verge Rangers)"),
			Condition: cmdr2B15AnOpponentControlsMoreLands,
		}},
	})
}
