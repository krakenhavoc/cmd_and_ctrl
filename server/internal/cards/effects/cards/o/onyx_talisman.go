package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Onyx Talisman — Artifact {2} (Ice Age, S82 OLD batch 847, #5636):
//
//	"Whenever a player casts a black spell, you may pay {3}. If you do, untap
//	 target permanent."
//
// The Ice Age Talisman cycle's shared Spec (OldB847TalismanSpec): any player's
// black spell triggers it, its controller chooses the permanent as the trigger goes
// on the stack and decides on the {3} as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.OldB847TalismanSpec("d19a19ed-33f9-400d-9669-c393d7127652", "Onyx Talisman", "B", "black"))
}
