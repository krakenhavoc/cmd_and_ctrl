package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lier, Disciple of the Drowned — Legendary Creature — Human Wizard
// {3}{U}{U}, 3/4:
//
//	"Spells can't be countered.
//	 Each instant and sorcery card in your graveyard has flashback. The
//	 flashback cost is equal to that card's mana cost."
//
// The first line is ADR 0106 §4's battlefield static (#1806) in its
// "any player" form: every player's spells, opponents' included.
//
// The second is a STANDING cast permission (ADR 0066), Underworld
// Breach's shape with Past in Flames' terms: derived off the
// battlefield, so a card that reaches the graveyard while Lier is out
// has flashback at once and every one loses it when Lier leaves. It
// names no Cost, so the flashback cost is the card's mana cost, and
// ExileOnResolution is flashback's CR 702.34a exile.
//
// ONE SIMPLIFICATION, the one Past in Flames already has: a card that
// prints its own flashback is offered only that one. CR 702.34 would let
// its owner pick either flashback cost; the printed one is kept, which
// can only cost the player more.
func init() {
	Register(Spec{
		OracleID:     "f224db3d-cdd2-43f2-b625-4b861efa6449",
		Name:         "Lier, Disciple of the Drowned",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"An instant or sorcery that already has flashback can only be cast for its own flashback cost, not for its mana cost.",
		},
		SpellsCantBeCountered: []game.CounterShieldStatic{
			AnyPlayersSpellsCantBeCountered("Spells can't be countered."),
		},
		CastPermissions: []game.CastPermission{{
			Zone:              game.ZoneGraveyard,
			Filter:            game.PermissionFilter{InstantOrSorceryOnly: true},
			AltCostKey:        "flashback",
			ExileOnResolution: true,
			Label:             "Flashback — its mana cost (Lier, Disciple of the Drowned)",
		}},
	})
}
