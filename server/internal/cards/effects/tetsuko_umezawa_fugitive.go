package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tetsuko Umezawa, Fugitive — Legendary Creature — Human Rogue {1}{U},
// 1/3:
//
//	"Creatures you control with power or toughness 1 or less can't be
//	 blocked."
//
// A block rule scoped to your creatures, with the power-or-toughness
// condition read live on the attacker each time a block is checked
// (so a creature pumped past 1 on both stats stops being unblockable
// at once). Tetsuko's own 1/3 qualifies, on her power.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ceeeacbc-01b0-4421-aaca-2ce6cdbe45d7",
		Name:         "Tetsuko Umezawa, Fugitive",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedWhile(ControlledBySourceController(), rfReprintBLowPowerOrToughness()),
		},
	})
}
