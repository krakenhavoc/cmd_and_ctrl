package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hematite Talisman — Artifact {2} (Ice Age, S82 OLD batch 847, #5636):
//
//	"Whenever a player casts a red spell, you may pay {3}. If you do, untap
//	 target permanent."
//
// The Ice Age Talisman cycle's shared Spec (OldB847TalismanSpec): any player's
// red spell triggers it, its controller chooses the permanent as the trigger goes
// on the stack and decides on the {3} as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.OldB847TalismanSpec("586601dd-ac73-4451-848d-f21e64feb6a9", "Hematite Talisman", "R", "red"))
}
