package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marauding Maulhorn — Creature — Beast {2}{R}{R}, 5/3 (Magic 2014, #5563):
//
//	"This creature attacks each combat if able unless you control a creature named Advocate of the Beast."
//
// Bloodrock Cyclops's requirement with Galactus's exemption: the requirement
// is off while its controller controls a creature named Advocate of the Beast
// (youControlACreatureNamed, read in the layer pass).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f4f0cd2-9777-4ea1-944d-306347dc48a2",
		Name:         "Marauding Maulhorn",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{m14MaraudingMaulhornMustAttack()},
	})
}
