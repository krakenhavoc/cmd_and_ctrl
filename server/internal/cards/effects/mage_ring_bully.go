package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mage-Ring Bully — Creature — Human Warrior {1}{R}, 2/2 (Magic Origins, #5549):
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
//	 This creature attacks each combat if able."
//
// Prowess is the engine keyword; the attack requirement is Juggernaut's
// (CR 508.1d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a0b4fa95-4c61-40ff-b645-a1dbe8825750",
		Name:            "Mage-Ring Bully",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Static:          []game.StaticAbility{AttacksEachCombat()},
	})
}
