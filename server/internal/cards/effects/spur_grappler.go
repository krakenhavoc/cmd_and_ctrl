package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spur Grappler — Creature — Beast {2}{R}, 2/1 (Prophecy, #5794):
//
//	"This creature gets +2/+1 as long as you control no untapped lands."
//
// Scoria Cat's shape at +2/+1: a layer 7c self pump (CR 613.4c) while its
// controller controls no untapped lands, read live (CR 611.3a,
// no_untapped_lands.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "42784364-719a-49ca-b515-64146a8f82e8",
		Name:         "Spur Grappler",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{SelfPumpWhileNoUntappedLands(2, 1)},
	})
}
