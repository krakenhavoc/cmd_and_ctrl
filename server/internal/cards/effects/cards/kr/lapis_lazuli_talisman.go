package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lapis Lazuli Talisman — Artifact {2} (Ice Age, S82 OLD batch 847, #5636):
//
//	"Whenever a player casts a blue spell, you may pay {3}. If you do, untap
//	 target permanent."
//
// The Ice Age Talisman cycle's shared Spec (OldB847TalismanSpec): any player's
// blue spell triggers it, its controller chooses the permanent as the trigger goes
// on the stack and decides on the {3} as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.OldB847TalismanSpec("83447e57-06f7-4e7b-91d4-7b8f71ea0676", "Lapis Lazuli Talisman", "U", "blue"))
}
