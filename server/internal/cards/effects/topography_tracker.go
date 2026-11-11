package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const topographyTrackerOracleID = "0171861e-98d5-4e58-834f-899d8ad7acea"

// Topography Tracker — Creature — Merfolk Scout {2}{G}, 2/2 (The Lost
// Caverns of Ixalan Commander, #5425):
//
//	"When this creature enters, create a Map token. (It's an artifact with
//	 "{1}, {T}, Sacrifice this token: Target creature you control explores.
//	 Activate only as a sorcery.")
//	 If a creature you control would explore, instead it explores, then it
//	 explores again."
//
// The Map is the catalog's token (CR 111.10s). The second line is a
// CR 614 replacement of the explore action (game.RegisterExploreAgainKey):
// every instruction for a creature you control to explore is two explores,
// the second starting once the first is complete (CR 701.44b), so a
// nonland card left on top is revealed again, and each one is an explore
// for "whenever a creature you control explores". The explore it adds is
// not replaced a second time by the same Tracker (CR 614.5), and a second
// Tracker applies once to what the first made: four explores. It is read
// through the catalog key, so a Tracker that lost its abilities replaces
// nothing.
//
// No simplification.
func init() {
	game.RegisterExploreAgainKey(topographyTrackerOracleID)
	Register(Spec{
		OracleID:     topographyTrackerOracleID,
		Name:         "Topography Tracker",
		Purpose:      game.Purpose{Tokens: 1},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Topography Tracker — create a Map token",
				Do(CreateToken{Template: MapToken(), N: 1})),
		},
	})
}
