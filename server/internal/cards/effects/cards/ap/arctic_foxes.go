package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arctic Foxes — Creature — Fox {1}{W}, 1/1 (Ice Age, S82 OLD, #5636):
//
//	"This creature can't be blocked by creatures with power 2 or greater as long as defending player
//	 controls a snow land."
//
// A block rule on the creature itself (CR 509.1b): the blocker's power is read live and "defending
// player" is the blocker's controller, who must control a snow land as blockers are declared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "898f0824-cf4c-4e34-8407-9a32fdae6dcb",
		Name:         "Arctic Foxes",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.S82OldB843DefenderControlsASnowLand, "creatures with power 2 or greater as long as defending player controls a snow land"),
		},
	})
}
