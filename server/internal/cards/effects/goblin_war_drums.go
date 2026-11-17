package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "29c21edd-781b-448e-824a-17bc8b8f4077",
		Name:         "Goblin War Drums",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{YoursOnly: true}, "menace"),
		},
	})
}
