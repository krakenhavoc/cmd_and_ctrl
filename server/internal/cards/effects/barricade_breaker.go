package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Barricade Breaker — Artifact Creature — Juggernaut {7}, 7/5:
//
//	"Improvise ... This creature attacks each combat if able."
//
// Improvise is Spec.TapCost; the attack requirement is AttacksEachCombat
// (Alexios, Deimos of Kosmos). No simplification.
func init() {
	Register(Spec{
		OracleID:     "93491d5c-8579-42ed-a700-e5c6f4238dea",
		Name:         "Barricade Breaker",
		Completeness: CompletenessFull,
		TapCost:      Improvise(),
		Static:       []game.StaticAbility{AttacksEachCombat()},
	})
}
