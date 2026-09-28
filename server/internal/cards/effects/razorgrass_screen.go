package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Razorgrass Screen — Artifact Creature — Wall, {1}, 2/1:
//
//	"Defender (This creature can't attack.)
//	 This creature blocks each combat if able."
//
// The requirement is on the Wall itself (#1597, CR 509.1c): its
// controller's pass in declare blockers is refused while it is untapped
// at home and some attacker it could legally block is coming at them.
// Defender is the enforced printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1818324c-2738-42e2-a78b-63abf325d8e9",
		Name:            "Razorgrass Screen",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static:          []game.StaticAbility{BlocksEachCombat()},
	})
}
