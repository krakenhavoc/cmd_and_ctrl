package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wizened Snitches — Creature — Faerie Rogue {3}{U}, 1/3 (Ravnica: City of
// Guilds, #5600):
//
//	"Flying
//	 Players play with the top card of their libraries revealed."
//
// Field of Dreams' revealed library tops (CR 401.5): every player's top
// card is public while it is on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "2ac4aca3-d5d6-4027-9f09-8b5f20e2438b",
		Name:              "Wizened Snitches",
		Completeness:      CompletenessFull,
		PrintedKeywords:   []string{"flying"},
		LibraryTopVisible: game.LibraryTopRevealedEachPlayer,
	})
}
