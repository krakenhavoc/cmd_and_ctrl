package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steel Leaf Champion — Creature — Elf Knight {G}{G}{G}, 5/4:
//
//	"This creature can't be blocked by creatures with power 2 or less."
//
// Riphook Raider's pair rule on the creature itself, read live as
// blockers are declared (CR 509.1b); once blocked it stays blocked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "666af637-9282-4121-962a-048e528f9221",
		Name:         "Steel Leaf Champion",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerLE(2), "creatures with power 2 or less"),
		},
	})
}
