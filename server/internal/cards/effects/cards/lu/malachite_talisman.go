package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Malachite Talisman — Artifact {2} (Ice Age, S82 OLD batch 847, #5636):
//
//	"Whenever a player casts a green spell, you may pay {3}. If you do, untap
//	 target permanent."
//
// The Ice Age Talisman cycle's shared Spec (OldB847TalismanSpec): any player's
// green spell triggers it, its controller chooses the permanent as the trigger goes
// on the stack and decides on the {3} as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.OldB847TalismanSpec("0be4877d-72ff-4951-aa52-f4f816abc844", "Malachite Talisman", "G", "green"))
}
