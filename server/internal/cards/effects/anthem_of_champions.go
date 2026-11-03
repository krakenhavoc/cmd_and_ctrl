package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anthem of Champions — Enchantment {G}{W}:
//
//	"Creatures you control get +1/+1."
//
// A layer 7c static over the creatures its controller controls
// (b16Anthem). The Anthem is not a creature, so no "other" guard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f1d8e9a6-1903-4be5-8609-1009a463f393",
		Name:         "Anthem of Champions",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{b16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller == source.Controller
		}, 1, 1)},
	})
}
