package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tattermunge Maniac — Creature — Goblin Warrior {R/G}, 2/1 (Shadowmoor,
// #5591):
//
//	"This creature attacks each combat if able."
//
// AttacksEachCombat, Amarant Coral's requirement (CR 508.1d): the attack
// declaration is refused while it could attack and stays home.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "69ba7864-0d09-4517-ad1a-7efaf437ded0",
		Name:         "Tattermunge Maniac",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{AttacksEachCombat()},
	})
}
