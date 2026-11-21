package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Root Sliver — Creature — Sliver {3}{G}, 2/2:
//
//	"This spell can't be countered.
//	 Sliver spells can't be countered."
//
// The rider is Spec.CantBeCountered; the second line is ADR 0106 §4's
// battlefield static (#1806) in its "any player" form: every player's
// Sliver spells. Subtype reads Card.HasSubtype, so a changeling spell is
// a Sliver spell (CR 702.73a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cb752313-dc7f-47fc-9077-9e3298e8f5fb",
		Name:            "Root Sliver",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		SpellsCantBeCountered: []game.CounterShieldStatic{
			effects.AnyPlayersSpellsCantBeCountered("Sliver spells can't be countered.", effects.Subtype("Sliver")),
		},
	})
}
