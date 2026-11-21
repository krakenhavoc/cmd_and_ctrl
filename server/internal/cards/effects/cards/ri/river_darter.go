package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// River Darter — Creature — Merfolk Warrior {2}{U}, 2/3 (Rivals of
// Ixalan, #5531):
//
//	"This creature can't be blocked by Dinosaurs."
//
// Juggernaut's "can't be blocked by Walls" with the creature type swapped:
// a blocker's types are read as blockers are declared (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "939a6815-bed1-43b0-bf4d-f0581ba0fed5",
		Name:         "River Darter",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.Subtype("Dinosaur"), "Dinosaurs"),
		},
	})
}
