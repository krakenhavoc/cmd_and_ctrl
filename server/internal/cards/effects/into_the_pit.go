package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Into the Pit — Enchantment {2}{B} (Duskmourn Commander):
//
//	"You may look at the top card of your library any time.
//	 You may cast spells from the top of your library by sacrificing a
//	 nonland permanent in addition to paying their other costs."
//
// Bolas's Citadel's top-of-library window (LibraryTopVisible and
// PlayFromTopOfYourLibrary), cast-only because a land is not a spell, with
// Exploration Broodship's additional sacrifice (CastPermission.
// SacrificeNonlandPermanentInAddition): the spell is cast for its mana cost
// and a nonland permanent you control, sacrificed as a cost (CR 118.8) and
// named when the cast is announced. A cast that names none is refused. The
// spell being cast is already on the stack, so it is never the sacrifice; Into
// the Pit itself may be.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "e330412f-46a1-4e7c-8766-f644606049e0",
		Name:              "Into the Pit",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{func() game.CastPermission {
			p := PlayFromTopOfYourLibrary(game.PermissionFilter{},
				"Cast a spell from the top of your library by sacrificing a nonland permanent in addition to paying its other costs (Into the Pit)")
			p.CastOnly = true
			p.AltCostKey = intoThePitAltCost
			p.SacrificeNonlandPermanentInAddition = true
			return p
		}()},
	})
}

// intoThePitAltCost is the key the sacrifice offer is claimed under. An
// on-disk identity: never renamed, never reused.
const intoThePitAltCost = "into_the_pit_sacrifice_nonland"
