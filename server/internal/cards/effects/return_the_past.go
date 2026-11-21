package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Return the Past — Enchantment {4}{R}{R} (Doctor Who, #5948):
//
//	"During your turn, each instant and sorcery card in your graveyard has
//	 flashback. Its flashback cost is equal to its mana cost."
//
// Lier, Disciple of the Drowned's permission (each instant and sorcery CARD
// in the graveyard, cast for its mana cost and exiled as it leaves the stack,
// CR 702.34a) restricted to its controller's own turn
// (TimingYourTurnOnly): on any other turn the cards have no flashback. The
// cast keeps the card's own timing, so an instant may be flashed back at
// instant speed during your turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "be2b3954-e4d7-438c-a549-252bd7f4a9bd",
		Name:         "Return the Past",
		Completeness: CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:              game.ZoneGraveyard,
			Filter:            game.PermissionFilter{CardInZone: true, InstantOrSorceryOnly: true},
			AltCostKey:        "flashback",
			ExileOnResolution: true,
			Timing:            game.TimingYourTurnOnly,
			Label:             "During your turn, flashback — its mana cost (Return the Past)",
		}},
	})
}
