package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hunted Ghoul — Creature — Zombie {B}, 1/2 (Avacyn Restored, S82 OLD, #5570):
//
//	"This creature can't block Humans."
//
// A block rule on the creature itself (CR 509.1b): it may be declared as a
// blocker of no attacker that is a Human, read through the layered view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "12827076-450d-42b1-98ca-9f9279a539fc",
		Name:         "Hunted Ghoul",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnSelf(), effects.OnMatching(effects.OfCreatureType("Human")), "this creature can't block Humans"),
		},
	})
}
