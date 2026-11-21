package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gobland — Land Creature — Mountain Goblin, 2/1 (Mystery Booster 2):
//
//	"(Gobland isn't a spell, it's affected by summoning sickness, and it has
//	 "{T}: Add {R}.")
//	 Gobland can't block."
//
// Dryad Arbor's shape: played as a land because its type line says Land,
// summoning sick because it is a creature whose mana ability taps (CR
// 302.6), and its Mountain subtype gives the {R}. "Can't block" is the
// restriction it prints on itself (Aesthir Glider's), which the block
// declaration checks (CR 509.1b). The mana ability is not declared: it is its
// Mountain subtype's (CR 305.6), which is why the catalog has no row for it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "74900080-80a6-4fd8-9557-5f773c0e4bec",
		Name:         "Gobland",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
	})
}
