package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clockwork Steed — Artifact Creature — Horse {4}, 0/3 (Homelands, S82 OLD b822,
// #5635):
//
//	"This creature enters with four +1/+0 counters on it.
//	This creature can't be blocked by artifact creatures.
//	At end of combat, if this creature attacked or blocked this combat, remove a
//	+1/+0 counter from it.
//	{X}, {T}: Put up to X +1/+0 counters on this creature. This ability can't
//	cause the total number of +1/+0 counters on this creature to be greater than
//	four. Activate only during your upkeep."
//
// OldB822ClockworkSpec, shared with Clockwork Swarm.
//
// No simplification.
func init() {
	effects.Register(effects.OldB822ClockworkSpec("1d6a1dab-8724-4b07-b109-c6ba2a7ea1ad", "Clockwork Steed",
		effects.And(effects.Artifact(), effects.Creature()), "artifact creatures"))
}
