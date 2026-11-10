package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Steel Leaf Champion — Creature — Elf Knight {G}{G}{G}, 5/4:
//
//	"This creature can't be blocked by creatures with power 2 or less."
//
// Riphook Raider's pair rule on the creature itself, read live as
// blockers are declared (CR 509.1b); once blocked it stays blocked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "666af637-9282-4121-962a-048e528f9221",
		Name:         "Steel Leaf Champion",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), PowerLE(2), "creatures with power 2 or less"),
		},
	})
}
