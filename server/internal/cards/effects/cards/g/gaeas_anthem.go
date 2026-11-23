package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gaea's Anthem — Enchantment {1}{G}{G} (Planar Chaos, #5595):
//
//	"Creatures you control get +1/+1."
//
// Anthem of Champions' layer 7c static over the creatures its controller
// controls (b16Anthem). The Anthem is not a creature, so no "other" guard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3754dce0-3e97-406f-8807-a4942a222c41",
		Name:         "Gaea's Anthem",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{effects.Labelled(effects.B16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller == source.Controller
		}, 1, 1), "Creatures you control get +1/+1.")},
	})
}
