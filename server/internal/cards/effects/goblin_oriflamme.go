package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Oriflamme — Enchantment {1}{R} (Foundations, #4465):
//
//	"Attacking creatures you control get +1/+0."
//
// Instigator Gang's front-face static (attackingCreaturesYouControlGetPower):
// a layer 7c boost over the creatures you control while they are attacking,
// so a creature gets it from the moment it is declared as an attacker until
// it leaves combat (CR 506.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "836bd011-2da5-443a-a814-19a664b98a1a",
		Name:         "Goblin Oriflamme",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{attackingCreaturesYouControlGetPower(1)},
	})
}
