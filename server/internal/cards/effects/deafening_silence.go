package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deafening Silence — {W} Enchantment (Special Guests):
//
//	"Each player can't cast more than one noncreature spell each turn."
//
// A cast restriction (CR 101.2) derived from the permanent being on the
// battlefield; it binds every player, its controller included, and a
// creature spell is never refused by it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6284ef01-d17d-4584-aecb-1b841ae638e5",
		Name:         "Deafening Silence",
		Completeness: CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			EachPlayerMaxNoncreatureSpellsPerTurn(1,
				"Deafening Silence — each player can't cast more than one noncreature spell each turn."),
		},
	})
}
