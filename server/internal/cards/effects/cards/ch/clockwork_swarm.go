package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clockwork Swarm — Artifact Creature — Insect {4}, 0/3 (Homelands, S82 OLD
// b822, #5635):
//
//	"This creature enters with four +1/+0 counters on it.
//	This creature can't be blocked by Walls.
//	At end of combat, if this creature attacked or blocked this combat, remove a
//	+1/+0 counter from it.
//	{X}, {T}: Put up to X +1/+0 counters on this creature. This ability can't
//	cause the total number of +1/+0 counters on this creature to be greater than
//	four. Activate only during your upkeep."
//
// OldB822ClockworkSpec, shared with Clockwork Steed.
//
// No simplification.
func init() {
	effects.Register(effects.OldB822ClockworkSpec("a6091c3e-b227-4e84-b75d-69512a0bddfb", "Clockwork Swarm",
		effects.Subtype("Wall"), "Walls"))
}
