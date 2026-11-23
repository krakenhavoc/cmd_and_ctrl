package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stony Silence — Enchantment {1}{W}:
//
//	"Activated abilities of artifacts can't be activated."
//
// The white Null Rod without the mana-ability exemption: it is Collector
// Ouphe's clause with nothing exempted, so a Sol Ring stops producing too.
// Symmetrical, and "artifacts" is the layered type line.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "12c63a45-7eb2-4262-b571-3358182f5610",
		Name:         "Stony Silence",
		Completeness: effects.CompletenessFull,
		ActivationRestrictions: []game.ActivationRestriction{
			effects.SourcesCantActivate(
				"Stony Silence — activated abilities of artifacts can't be activated.",
				effects.Artifact(), false),
		},
	})
}
