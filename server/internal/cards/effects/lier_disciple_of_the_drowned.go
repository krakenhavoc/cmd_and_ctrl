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
// A card that prints its own flashback has two flashback abilities
// under Lier (CR 113.2c), and its owner casts it with either (CR
// 702.34a): Faithless Looting for its printed {2}{R} or for its {R} mana
// cost. The printed one is claimed as "flashback", Lier's as
// "flashback/granted" (game.GrantedInstanceKey, #3938, ADR 0066
// decision 3 as amended 2026-10-16), and the cast picker lists both.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f224db3d-cdd2-43f2-b625-4b861efa6449",
		Name:         "Lier, Disciple of the Drowned",
		Completeness: CompletenessFull,
		SpellsCantBeCountered: []game.CounterShieldStatic{
			AnyPlayersSpellsCantBeCountered("Spells can't be countered."),
		},
		CastPermissions: []game.CastPermission{{
			Zone: game.ZoneGraveyard,
			// "Each instant and sorcery CARD in your graveyard": the card
			// in the graveyard, not the spell (CR 715.4, 712.8a; #4509).
			Filter:            game.PermissionFilter{CardInZone: true, InstantOrSorceryOnly: true},
			AltCostKey:        "flashback",
			ExileOnResolution: true,
			Label:             "Flashback — its mana cost (Lier, Disciple of the Drowned)",
		}},
	})
}
