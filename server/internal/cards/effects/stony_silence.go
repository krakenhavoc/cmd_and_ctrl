package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "12c63a45-7eb2-4262-b571-3358182f5610",
		Name:         "Stony Silence",
		Completeness: CompletenessFull,
		ActivationRestrictions: []game.ActivationRestriction{
			SourcesCantActivate(
				"Stony Silence — activated abilities of artifacts can't be activated.",
				Artifact(), false),
		},
	})
}
