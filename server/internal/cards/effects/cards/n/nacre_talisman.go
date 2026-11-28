package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nacre Talisman — Artifact {2} (Ice Age, S82 OLD batch 847, #5636):
//
//	"Whenever a player casts a white spell, you may pay {3}. If you do, untap
//	 target permanent."
//
// The Ice Age Talisman cycle's shared Spec (OldB847TalismanSpec): any player's
// white spell triggers it, its controller chooses the permanent as the trigger goes
// on the stack and decides on the {3} as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.OldB847TalismanSpec("f87aaed9-7f78-4f8b-8af5-44d4be1d6f30", "Nacre Talisman", "W", "white"))
}
