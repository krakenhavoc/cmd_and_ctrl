package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "6284ef01-d17d-4584-aecb-1b841ae638e5",
		Name:         "Deafening Silence",
		Completeness: effects.CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			effects.EachPlayerMaxNoncreatureSpellsPerTurn(1,
				"Deafening Silence — each player can't cast more than one noncreature spell each turn."),
		},
	})
}
