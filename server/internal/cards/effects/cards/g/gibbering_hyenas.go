package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gibbering Hyenas — Creature — Hyena {2}{G}, 3/2 (Mirage, S82 OLD batch 810,
// #5633):
//
//	"This creature can't block black creatures."
//
// Hinterland Drake's blocker-side rule with the other attacker set: the pair is
// refused when the attacker is black (CR 509.1b). Colour is read as it stands
// in the layered view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a073f00-0d03-43d0-a0a8-f5bcee0855fc",
		Name:         "Gibbering Hyenas",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnSelf(), effects.OnMatching(effects.And(effects.Creature(), effects.OfColor("B"))), "it can't block black creatures"),
		},
	})
}
