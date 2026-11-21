package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin War Drums — Enchantment {2}{R} (Fallen Empires, #5637):
//
//	"Creatures you control have menace. (They can't be blocked except by
//	 two or more creatures.)"
//
// Fervor's anthem with menace: a layer-6 grant to every creature its
// controller controls, through TribalKeywordGrant with no tribe named.
// The block check reads the effective keyword, so a creature that
// arrives later is covered at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "29c21edd-781b-448e-824a-17bc8b8f4077",
		Name:         "Goblin War Drums",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{YoursOnly: true}, "menace"),
		},
	})
}
