package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rampart Crawler — Creature — Lizard Mercenary {B}, 1/1 (Mercadian Masques, S82 OLD batch 687, #5619):
//
//	"This creature can't be blocked by Walls."
//
// Juggernaut's "can't be blocked by Walls": a blocker's types are read as blockers are declared (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "302c0400-043d-42fa-abbb-488e520116ba",
		Name:         "Rampart Crawler",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.Subtype("Wall"), "Walls"),
		},
	})
}
