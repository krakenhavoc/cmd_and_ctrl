package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prized Unicorn — Creature — Unicorn, {3}{G}, 2/2:
//
//	"All creatures able to block this creature do so."
//
// Lure printed on the creature itself (#1597, CR 509.1c). It is the
// Unicorn's own ability, so a Unicorn that loses all abilities stops
// luring (CR 613.1f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "766787ef-654a-4544-b5d7-d2f96949edf1",
		Name:         "Prized Unicorn",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{AllAbleToBlockDoSo()},
	})
}
